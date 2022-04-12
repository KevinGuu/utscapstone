package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type ServerParameters struct {
	port     int    // webhook server port
	certFile string // path to the x509 certificate for https
	keyFile  string // path to the x509 private key matching `CertFile`
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var parameters ServerParameters

var (
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
)

var config *rest.Config
var clientSet *kubernetes.Clientset

func main() {
	useKubeConfig := os.Getenv("USE_KUBECONFIG")
	kubeConfigFilePath := os.Getenv("KUBECONFIG")

	flag.IntVar(&parameters.port, "port", 8443, "Webhook server port.")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	if len(useKubeConfig) == 0 {
		// default to service account in cluster token
		c, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = c
	} else {
		//load from a kube config
		var kubeconfig string

		if kubeConfigFilePath == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		} else {
			kubeconfig = kubeConfigFilePath
		}

		fmt.Println("kubeconfig: " + kubeconfig)

		c, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		config = c
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	clientSet = cs

	pods, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	fmt.Println("Now listening on the /mutate endpoint")
	http.HandleFunc("/mutate", HandleMutate)
	log.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(parameters.port), parameters.certFile, parameters.keyFile, nil))
}

func HandleMutate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("----- Incoming /mutate request")

	// read request and write it to /tmp/request
	body, _ := ioutil.ReadAll(r.Body)
	err := ioutil.WriteFile("/tmp/request", body, 0644)
	if err != nil {
		panic(err.Error())
	}

	// demarshal requset to AdmissionReview object, handle errors
	var admissionReviewReq v1beta1.AdmissionReview
	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Errorf("Could not deserialize request: %v", err)
	} else if admissionReviewReq.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		errors.New("Malformed admission review: request is nil")
	}

	// print metadata to stdout
	fmt.Println("Request Kind:", admissionReviewReq.Request.Kind)
	fmt.Println("Request Operation:", admissionReviewReq.Request.Operation)
	fmt.Println("Request Name:", admissionReviewReq.Request.Name)
	fmt.Println("Request Namespace:", admissionReviewReq.Request.Namespace)

	// get NS labels
	ns := admissionReviewReq.Request.Namespace
	ptrNs, err := clientSet.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
	fmt.Println("Namespace name:", ptrNs.ObjectMeta.Name)
	annotations := ptrNs.ObjectMeta.Annotations
	for k, v := range annotations {
		fmt.Println(k, v)
	}

	// var pod apiv1.Pod

	// err = json.Unmarshal(admissionReviewReq.Request.Object.Raw, &pod)

	// if err != nil {
	// 	fmt.Errorf("Could not unmarshal pod on admission request: %v", err)
	// }

	// var patches []patchOperation

	// labels := pod.ObjectMeta.Labels
	// labels["example-webhook"] = "it-worked"

	// patches = append(patches, patchOperation{
	// 	Op:    "add",
	// 	Path:  "/metadata/labels",
	// 	Value: labels,
	// })

	// patchBytes, err := json.Marshal(patches)

	// if err != nil {
	// 	fmt.Errorf("could not marshal JSON patch: %v", err)
	// }

	// admissionReviewResponse := v1beta1.AdmissionReview{
	// 	Response: &v1beta1.AdmissionResponse{
	// 		UID:     admissionReviewReq.Request.UID,
	// 		Allowed: true,
	// 	},
	// }

	// admissionReviewResponse.Response.Patch = patchBytes

	// bytes, err := json.Marshal(&admissionReviewResponse)
	// if err != nil {
	// 	fmt.Errorf("marshaling response: %v", err)
	// }

	// w.Write(bytes)

}

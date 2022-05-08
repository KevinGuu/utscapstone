package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	yaml "gopkg.in/yaml.v2"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
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

type Config struct {
	Container []corev1.Container `yaml:"initContainers"`
}

var parameters ServerParameters

var sidecarConfigFile string

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
	flag.StringVar(&sidecarConfigFile, "sidecar-config-file", "/etc/webhook/config/sidecarconfig.yaml", "Sidecar injector configuration file.")
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

	// pods, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	// if err != nil {
	// 	panic(err.Error())
	// }
	// fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	fmt.Println("Now listening on the /mutate endpoint")
	http.HandleFunc("/mutate", HandleMutate)
	http.ListenAndServeTLS(":"+strconv.Itoa(parameters.port), parameters.certFile, parameters.keyFile, nil)
}

func HandleMutate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("----- Incoming request to /mutate")

	// read request
	body, _ := ioutil.ReadAll(r.Body)

	err := ioutil.WriteFile("/tmp/request", body, 0644)
	if err != nil {
		panic(err.Error())
	}

	// demarshal requset to AdmissionReview object, handle errors
	var admissionReview admissionv1.AdmissionReview
	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReview); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Errorf("Could not deserialize request: %v", err)
	} else if admissionReview.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		errors.New("Malformed admission review: request is nil")
	}
	fmt.Println("Demarshaled request to AdmissionReview object")

	// print Request metadata to stdout
	fmt.Println("Request Kind:", admissionReview.Request.Kind, "Request Operation:", admissionReview.Request.Operation, "Request Name:", admissionReview.Request.Name, "Request Namespace:", admissionReview.Request.Namespace)

	// get NS labels, check if cidr-range in labels list, if so, parse and get range
	ns := admissionReview.Request.Namespace
	ptrNs, err := clientSet.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	annotations := ptrNs.ObjectMeta.Annotations
	if v, found := annotations["cidr-range"]; found {
		fmt.Println("Found cidr-range in annotations, range is: ", v)
	} else {
		panic(err.Error())
	}

	// load sidecar config from mounted file
	sidecarConfig, err := loadConfig(sidecarConfigFile)
	fmt.Println("Loaded sidecar config")

	// set env var
	sidecarConfig.Container[0].Env[0].Value = annotations["cidr-range"]

	// set imagepullpolicy
	pp := corev1.PullPolicy("Always")
	sidecarConfig.Container[0].ImagePullPolicy = pp

	// add Linux capabilities
	sc := corev1.SecurityContext{}
	c := corev1.Capabilities{}
	c.Add = []corev1.Capability{"NET_ADMIN", "NET_RAW"}
	sc.Capabilities = &c
	isPrivileged := bool(true)
	sc.Privileged = &isPrivileged
	sidecarConfig.Container[0].SecurityContext = &sc
	fmt.Println(sidecarConfig)

	// create and apply container injection patch
	var patches []patchOperation

	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  "/spec/initContainers",
		Value: sidecarConfig.Container,
	})

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		fmt.Errorf("could not marshal JSON patch: %v", err)
	}

	admissionReviewResponse := admissionv1.AdmissionReview{
		Response: &admissionv1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: true,
		},
	}

	admissionReviewResponse.Response.Patch = patchBytes
	bytes, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		fmt.Errorf("marshaling response: %v", err)
	}

	w.Write(bytes)
}

func loadConfig(configFile string) (*Config, error) {
	// read configfile from flag var into data as bytes
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// unmarshal data bytes into config struct
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

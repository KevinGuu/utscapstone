import iptc
import os


def main():
    # read in CIDR ranges from env vars
    ip_range_ns = os.environ["ip_range_ns"]
    ip_range_cluster = os.environ["ip_range_cluster"]
    print(f"ip_range_ns: {ip_range_ns} ip_range_cluster {ip_range_cluster}")

    # setup input/output policy chains
    chain_input = iptc.Chain(iptc.Table(iptc.Table.FILTER), "INPUT")
    chain_output = iptc.Chain(iptc.Table(iptc.Table.FILTER), "OUTPUT")

    # clear config rules
    print("Flushing existing config...")
    chain_input.flush()
    chain_output.flush()
    print("Existing config flushed")

    # create INPUT/OUTPUT rules
    print("Setting up iptable rules...")
    rule = iptc.Rule()
    target_accpet = rule.create_target("ACCEPT")
    rule.src = ip_range_ns
    rule.dst = ip_range_ns
    rule.target = target_accpet
    chain_input.insert_rule(rule)
    chain_output.insert_rule(rule)
    print("Finished setting up iptables rules")


if __name__ == "__main__":
    main()

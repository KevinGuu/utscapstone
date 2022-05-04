#!/usr/bin/env python3
import iptc
import os
from time import sleep

def main():
    # read in CIDR ranges from env vars
    ip_range_ns = os.environ["ip_range_ns"]
    ip_range_cluster = os.environ["ip_range_cluster"]
    print(f"ip_range_ns: {ip_range_ns} ip_range_cluster {ip_range_cluster}")

    # setup input/output policy chains
    table = iptc.Table(iptc.Table.FILTER)
    chain_input = iptc.Chain(table, "INPUT")
    chain_output = iptc.Chain(table, "OUTPUT")

    # clear existing rules
    print("Flushing existing config...")
    chain_input.flush()
    chain_output.flush()
    print("Existing config flushed")

    # create INPUT/OUTPUT rules
    print("Setting up iptable rules...")
    rule = iptc.Rule()
    target_accpet = rule.create_target("DROP")
    rule.src = ip_range_cluster
    rule.dst = ip_range_cluster
    rule.target = target_accpet
    chain_input.insert_rule(rule)
    chain_output.insert_rule(rule)

    rule = iptc.Rule()
    target_accpet = rule.create_target("ACCEPT")
    rule.src = ip_range_ns
    rule.dst = ip_range_ns
    rule.target = target_accpet
    chain_input.insert_rule(rule)
    chain_output.insert_rule(rule)
    print("Finished setting up iptables rules")

    print("Input chain rules:")
    for rule in chain_input.rules:
        print(f"----- Proto: {rule.protocol} src: {rule.src} dst: {rule.dst} target: {rule.target.name}")
    print("Output chain rules:")
    for rule in chain_output.rules:
        print(f"----- Proto: {rule.protocol} src: {rule.src} dst: {rule.dst} target: {rule.target.name}")

    sleep(60*60*24)


if __name__ == "__main__":
    main()

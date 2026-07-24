# Sketch a three-tier VPC

Before CloudFormation, draw the network: public edge, private compute, NAT for egress, and security groups that only open what is required.

## Tasks

1. Update `/workspace/vpc/network.yaml`:
   - VPC CIDR `10.0.0.0/16`
   - `public-a` = `10.0.0.0/24`
   - `private-a` = `10.0.10.0/24`
   - `nat_gateway_subnet: public-a`
2. Write `/workspace/vpc/security-groups.yaml`:
   - `public-lb` allows TCP/443 from `0.0.0.0/0`
   - `private-api` allows TCP/8080 only from `public-lb`

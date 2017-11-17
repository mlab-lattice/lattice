###############################################################################
# Variables

variable "availability_zones" {
  description = "List of Availability Zones to use"
  type        = "list"
}

###############################################################################
# Provider

provider "aws" {
  region = "us-east-1"
}

###############################################################################
# Networking

# VPC
resource "aws_vpc" "vpc" {
  cidr_block           = "10.240.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags {
    Name = "lattice packer"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "igw" {
  vpc_id = "${aws_vpc.vpc.id}"

  tags {
    Name = "lattice packer"
  }
}

#Route Table
resource "aws_route_table" "route_table" {
  vpc_id = "${aws_vpc.vpc.id}"

  tags {
    Name = "lattice packer"
  }
}

# Route for Internet Gateway
resource "aws_route" "igw" {
  route_table_id         = "${aws_route_table.route_table.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.igw.id}"
}

# Subnets
resource "aws_subnet" "subnet" {
  vpc_id            = "${aws_vpc.vpc.id}"
  availability_zone = "${element(var.availability_zones, 0)}"
  cidr_block        = "${cidrsubnet(aws_vpc.vpc.cidr_block, 2, count.index)}"
}

# Associate Subnets with Route Table
resource "aws_route_table_association" "route_table_association" {
  subnet_id      = "${aws_subnet.subnet.id}"
  route_table_id = "${aws_route_table.route_table.id}"
}

###############################################################################
# Output

output "vpc_id" {
  value = "${aws_vpc.vpc.id}"
}

output "subnet_id" {
  value = "${aws_subnet.subnet.id}"
}

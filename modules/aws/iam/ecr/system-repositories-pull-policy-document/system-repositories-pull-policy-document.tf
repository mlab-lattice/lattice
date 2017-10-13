###############################################################################
# Variables

variable "aws_account_id" {}
variable "region" {}
variable "system_id" {}

###############################################################################
# Output

output "json" {
  value = "${module.ecr_repository_pull_policy_document.json}"
}

###############################################################################
# IAM

module "ecr_repository_pull_policy_document" {
  source = "../repository-pull-policy-document"

  aws_account_id = "${var.aws_account_id}"
  region         = "${var.region}"
  repository     = "lattice/systems/${var.system_id}/*"
}

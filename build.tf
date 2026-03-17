locals {
  go_src_dir       = "${path.module}/files/src"
  go_binary_output = "${path.module}/tf_generated/shared/bootstrap"
  go_zip_output    = "${path.module}/tf_generated/shared/bootstrap.zip"
  is_linux_build   = data.uname.build_host.operating_system != "windows"

  go_src_hash = sha1(join("", [
    for f in fileset("${path.module}/files/src", "**") :
    filesha1("${path.module}/files/src/${f}")
  ]))

  abs_go_src        = replace(abspath(local.go_src_dir), "\\", "/")
  abs_go_binary_out = replace(abspath(local.go_binary_output), "\\", "/")
}

data "uname" "build_host" {}

# Runs during plan phase so the zip exists before filebase64sha256 is called
# by child modules. Rebuilds only when source hash changes (cached via .srchash file).
data "external" "shared_go_build" {
  program = local.is_linux_build ? ["bash", "${path.module}/scripts/build.sh"] : ["powershell", "-File", "${path.module}/scripts/build.ps1"]

  query = {
    src_dir    = local.abs_go_src
    binary_out = local.abs_go_binary_out
    src_hash   = local.go_src_hash
  }
}

data "archive_file" "shared_lambda_zip" {
  type        = "zip"
  source_file = local.go_binary_output
  output_path = local.go_zip_output

  depends_on = [data.external.shared_go_build]
}

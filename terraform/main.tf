locals {
  src_path    = abspath("../src/")
  db_file     = abspath("../GeoLite2-City.mmdb")
  zip_file    = abspath("../gomercury.zip")
}

data "archive_file" "gomercury" {
  type        = "zip"
  source_dir  = local.src_path
  output_path = local.zip_file
}

resource "google_storage_bucket" "bucket" {
  project     = "gomercury-356415"
  name        = "gomercury-bucket356415"
  location    = "US"
}

resource "google_storage_bucket_object" "archive" {
  name   = "gomercury.zip"
  bucket = google_storage_bucket.bucket.name
  source = data.archive_file.gomercury.output_path
}

resource "google_storage_bucket_object" "file" {
  name   = "GeoLite2-City.mmdb"
  bucket = google_storage_bucket.bucket.name
  source = local.db_file
}

resource "google_storage_bucket_access_control" "public_rule" {
  bucket = google_storage_bucket.bucket.name
  role   = "READER"
  entity = "allUsers"
}

resource "google_cloudfunctions_function" "function" {
  name                  = "GoMercury"
  description           = "A simple Golang API that provides information related to IP addresses"
  runtime               = "go116"
  project               = "gomercury-356415"
  region                = "us-central1"
  available_memory_mb   = 256
  min_instances         = 1
  max_instances         = 10
  source_archive_bucket = google_storage_bucket.bucket.name
  source_archive_object = google_storage_bucket_object.archive.name
  trigger_http          = true
}

# IAM entry for all users to invoke the function
resource "google_cloudfunctions_function_iam_member" "invoker" {
  project        = google_cloudfunctions_function.function.project
  region         = google_cloudfunctions_function.function.region
  cloud_function = google_cloudfunctions_function.function.name

  role   = "roles/cloudfunctions.invoker"
  member = "allUsers"
}
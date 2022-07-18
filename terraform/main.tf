resource "google_storage_bucket" "bucket" {
  project        = "gomercury-356415"
  name     = "gomercury-bucket356415"
  location = "US"
}

resource "google_storage_bucket_object" "archive" {
  name   = "gomercury.zip"
  bucket = google_storage_bucket.bucket.name
  source = "../gomercury.zip"
}

resource "google_storage_bucket_object" "file" {
  name   = "GeoLite2-City.mmdb"
  bucket = google_storage_bucket.bucket.name
  source = "../GeoLite2-City.mmdb"
}

resource "google_storage_bucket_access_control" "public_rule" {
  bucket = google_storage_bucket.bucket.name
  role   = "READER"
  entity = "allUsers"
}

# This resource will destroy (potentially immediately) after null_resource.next
resource "null_resource" "previous" {}

resource "time_sleep" "wait_30_seconds" {
  depends_on = [null_resource.previous]

  create_duration = "30s"
}

resource "google_cloudfunctions_function" "function" {
  name                  = "GoMercury"
  description           = "A simple Golang API that provides information related to IP addresses"
  runtime               = "go116"
  project               = "gomercury-356415"
  region                = "us-central1"
  available_memory_mb   = 256
  source_archive_bucket = google_storage_bucket.bucket.name
  source_archive_object = google_storage_bucket_object.archive.name
  trigger_http          = true
  depends_on = [time_sleep.wait_30_seconds]
}

# IAM entry for all users to invoke the function
resource "google_cloudfunctions_function_iam_member" "invoker" {
  project        = google_cloudfunctions_function.function.project
  region         = google_cloudfunctions_function.function.region
  cloud_function = google_cloudfunctions_function.function.name

  role   = "roles/cloudfunctions.invoker"
  member = "allUsers"
}
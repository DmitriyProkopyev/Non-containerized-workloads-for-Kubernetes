job "test-workload" {
  datacenters = ["dc1"]
  group "main" {
    count = 2

    task "main" {
      driver = "docker"
      config {
        image = "nginx:latest"
      }

      resources {
        cpu    = 1000
        memory = 1024
      }
    }
  }
}
job "mongo" {
  datacenters = ["dc1"]

  group "mongo-group" {
    count = 1

    # Локальный хост-том
    volume "mongo-data" {
      type      = "host"
      source    = "mongo-data" # ссылка на client.host_volume
      read_only = false
    }

    task "mongodb" {
      driver = "exec"

      config {
        command = "mongod"
        args = [
          "--dbpath=/opt/mongo/data",
          "--bind_ip_all"
        ]
      }

      resources {
        cpu    = 500
        memory = 512

        network {
          port "db" {
            static = 27017
          }
        }
      }

      volume_mount {
        volume      = "mongo-data"
        destination = "/opt/mongo/data"
        read_only   = false
      }
    }
  }
}
# Общие настройки
data_dir = "/opt/nomad/data"
bind_addr = "0.0.0.0"

# Сервер Nomad
server {
  enabled          = true
  bootstrap_expect = 1
}

# Клиент Nomad
client {
  enabled = true

  # Описание хост-тома для MongoDB
  host_volume "mongo-data" {
    path      = "/opt/nomad/mongo-data"
    read_only = false
  }

  # # Описание хост-тома для Kafka
  # host_volume "kafka" {
  #   path      = "/opt/kafka"
  #   read_only = false
  # }
}

# Адреса для подключения (укажи внешний IP машины)
advertise {
  http = "192.168.1.113:4646"
  rpc  = "192.168.1.113:4647"
  serf = "192.168.1.113:4648"
}

package main

import (
    "log"
    "net/http"
    "net/url"
    "strings"
    "time"
    "fmt"
    "os"

    "github.com/go-redis/redis/v8"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "golang.org/x/net/context"
)


type RedisExporter struct {
    clients   map[string]*redis.Client
    upMetrics *prometheus.GaugeVec
}


func NewRedisExporter(redisAddresses []string) *RedisExporter {
    clients := make(map[string]*redis.Client)

    for _, addr := range redisAddresses {
        if !strings.HasPrefix(addr, "redis://") {
            addr = fmt.Sprintf("redis://%s", addr)
        }

        u, err := url.Parse(addr)
        if err != nil {
            log.Fatalf("Failed to parse Redis address %s: %v", addr, err)
        }

        log.Printf("Redis info is %s: %v\n", u)

        password, ok := u.User.Password()
        if !ok {
            log.Fatalf("Password not found for Redis address %s", addr)
        }
        log.Printf("Connecting to Redis at %s with password %s", u.Host, password)

        client := redis.NewClient(&redis.Options{
            Addr:     u.Host,
            Password: password,
        })
        clients[u.Host] = client
    }

    upMetrics := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "redis_instance_up",
            Help: "Shows if a Redis instance is up (1) or down (0).",
        },
        []string{"address"},
    )

    prometheus.MustRegister(upMetrics)

    return &RedisExporter{
        clients:   clients,
        upMetrics: upMetrics,
    }
}


func (re *RedisExporter) CollectMetrics() {
    ctx := context.Background()
    for addr, client := range re.clients {
        _, err := client.Ping(ctx).Result()
        if err != nil {
            re.upMetrics.WithLabelValues(addr).Set(0)
            log.Printf("Redis at %s is down: %v\n", addr, err)
        } else {
            re.upMetrics.WithLabelValues(addr).Set(1)
            log.Printf("Redis at %s is up\n", addr)
        }
    }
}


func main() {
    redisAddresses := os.Getenv("REDIS_NODES")
    port := os.Getenv("EXPOSE_PORT")

    if redisAddresses == "" {
        log.Fatal("No Redis addresses provided. Use REDIS_NODES environment variable.")
    }

    addressList := strings.Split(redisAddresses, ",")

    exporter := NewRedisExporter(addressList) // Only pass the address list

    go func() {
        for {
            exporter.CollectMetrics()
            time.Sleep(10 * time.Second)
        }
    }()

    http.Handle("/metrics", promhttp.Handler())
    address := fmt.Sprintf(":%s", port)
    log.Printf("Starting Redis Exporter on %s", address)
    if err := http.ListenAndServe(address, nil); err != nil {
        log.Fatalf("Failed to start HTTP server: %v", err)
    }
}

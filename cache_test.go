package cache_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/redis.v4"
	"gopkg.in/vmihailenco/msgpack.v2"

	"gopkg.in/go-redis/cache.v4"
	"gopkg.in/go-redis/cache.v4/lrucache"
)

func TestModels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cache")
}

var _ = Describe("Codec", func() {
	var ring *redis.Ring
	var codec *cache.Codec

	testCodec := func() {
		It("Gets and Sets data", func() {
			key := "mykey"
			obj := Object{
				Str: "mystring",
				Num: 42,
			}

			err := codec.Set(&cache.Item{
				Key:        key,
				Object:     obj,
				Expiration: time.Hour,
			})
			Expect(err).NotTo(HaveOccurred())

			var wanted Object
			err = codec.Get(key, &wanted)
			Expect(err).NotTo(HaveOccurred())
			Expect(wanted).To(Equal(obj))
		})

		It("Gets and Sets nil", func() {
			key := "mykey"

			err := codec.Set(&cache.Item{
				Key:        key,
				Expiration: time.Hour,
			})
			Expect(err).NotTo(HaveOccurred())

			err = codec.Get(key, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Deletes key", func() {
			key := "mykey"

			err := codec.Set(&cache.Item{
				Key:        key,
				Expiration: time.Hour,
			})
			Expect(err).NotTo(HaveOccurred())

			err = codec.Delete(key)
			Expect(err).NotTo(HaveOccurred())

			err = codec.Get(key, nil)
			Expect(err).To(Equal(cache.ErrCacheMiss))
		})
	}

	BeforeEach(func() {
		ring = redis.NewRing(&redis.RingOptions{
			Addrs: map[string]string{
				"server1": ":6379",
				"server2": ":6380",
			},

			DialTimeout:  3 * time.Second,
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
		})
	})

	Context("without Cache", func() {
		BeforeEach(func() {
			codec = &cache.Codec{
				Redis: ring,

				Marshal: func(v interface{}) ([]byte, error) {
					return msgpack.Marshal(v)
				},
				Unmarshal: func(b []byte, v interface{}) error {
					return msgpack.Unmarshal(b, v)
				},
			}
		})

		testCodec()
	})

	Context("with Cache", func() {
		BeforeEach(func() {
			codec = &cache.Codec{
				Redis:      ring,
				LocalCache: lrucache.New(time.Minute, 1000),

				Marshal: func(v interface{}) ([]byte, error) {
					return msgpack.Marshal(v)
				},
				Unmarshal: func(b []byte, v interface{}) error {
					return msgpack.Unmarshal(b, v)
				},
			}
		})

		testCodec()
	})
})

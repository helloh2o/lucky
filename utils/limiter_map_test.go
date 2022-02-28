package utils

import (
	"github.com/helloh2o/lucky/log"
	"testing"
	"time"
)

func TestLimiterMap_IsLimited(t *testing.T) {
	log.New("release", "", 0)
	// 每秒只有一个 /fff 请求可以被接收
	for {
		log.Release("================================================")
		go func() {
			log.Release("%t", Limiter.IsLimited("/ffff", 1))
			time.Sleep(time.Second)
		}()
		go func() {
			log.Release("%t", Limiter.IsLimited("/ffff", 1))
			time.Sleep(time.Second)
		}()
		go func() {
			log.Release("%t", Limiter.IsLimited("/ffff", 1))
			time.Sleep(time.Second)
		}()
		go func() {
			log.Release("%t", Limiter.IsLimited("/ffff", 1))
			time.Sleep(time.Second)
		}()
		time.Sleep(time.Second)
	}
}

func TestLimiterMap_IsV2Limited(t *testing.T) {
	log.New("release", "", 0)
	duration := time.Second * 5
	max := int64(3)
	//log.Release("%t", Limiter.IsV2Limited("/ffff", duration))
	// 每duration只有一个 /fff 请求可以被接收
	for {
		log.Release("================================================")
		go func() {
			ok, n := Limiter.IsV2Limited("/ffff", duration, max)
			log.Release("%t-> %d", ok, n)
		}()
		go func() {
			ok, n := Limiter.IsV2Limited("/ffff", duration, max)
			log.Release("%t-> %d", ok, n)
		}()
		go func() {
			ok, n := Limiter.IsV2Limited("/ffff", duration, max)
			log.Release("%t-> %d", ok, n)
		}()
		go func() {
			ok, n := Limiter.IsV2Limited("/ffff", duration, max)
			log.Release("%t-> %d", ok, n)
		}()
		go func() {
			ok, n := Limiter.IsV2Limited("/ffff", duration, max)
			log.Release("%t-> %d", ok, n)
		}()
		go func() {
			ok, n := Limiter.IsV2Limited("/ffff", duration, max)
			log.Release("%t-> %d", ok, n)
		}()
		time.Sleep(time.Second)
	}
}

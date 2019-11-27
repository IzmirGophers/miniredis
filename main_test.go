package main

import (
	"testing"

	"github.com/cornelk/hashmap"
)

func init() {
	store = &hashmap.HashMap{}
	store.Set("foo", "bar")
	store.Set("foo1", "bar")
	store.Set("foo2", "bar")

}

func TestGet(t *testing.T) {
	rsp := get(store, []string{"GET", "foo"})
	if len(rsp) == 1 && rsp[0] != "bar" {
		t.Fail()
	}
}

func TestDBSize(t *testing.T) {
	rsp := dbSize(store, []string{"DBSIZE"})
	if len(rsp) == 1 && rsp[0] != "3" {
		t.Fail()
	}
}

func TestKeys(t *testing.T) {
	rsp := keys(store, []string{"KEYS"})
	if len(rsp) == 3 && rsp[0] != "foo" && rsp[1] != "foo1" && rsp[2] != "foo2" {
		t.Fail()
	}
}

func TestSet(t *testing.T) {
	rsp := set(store, []string{"SET", "foo3", "foo"})
	if len(rsp) == 1 && rsp[0] != responseOK {
		t.Fail()
	}
}

func TestMset(t *testing.T) {
	rsp := mset(store, []string{"MSET", "foo4", "bar", "foo5", "bar"})
	if len(rsp) == 1 && rsp[0] != responseOK {
		t.Fail()
	}
}

func TestMget(t *testing.T) {
	rsp := mget(store, []string{"MGET", "foo", "foo1", "foo2"})
	if len(rsp) == 3 && rsp[0] != "bar" && rsp[1] != "bar" && rsp[2] != "bar" {
		t.Fail()
	}
}

func BenchmarkGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		get(store, []string{"GET", "foo"})
	}
}

func BenchmarkSet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		set(store, []string{"GET", "bar", "foo"})
	}
}

func BenchmarkMGet(b *testing.B) {
	for n := 0; n < b.N; n++ {
		mget(store, []string{"MGET", "foo", "bar"})
	}
}

func BenchmarkMset(b *testing.B) {
	for n := 0; n < b.N; n++ {
		mset(store, []string{"MGET", "foo1", "bar1", "foo2", "bar3"})
	}
}

func BenchmarkKeys(b *testing.B) {
	for n := 0; n < b.N; n++ {
		keys(store, []string{"KEYS"})
	}
}

func BenchmarkDBSize(b *testing.B) {
	for n := 0; n < b.N; n++ {
		dbSize(store, []string{"DBSIZE"})
	}
}

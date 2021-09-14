package dlock

import (
	"context"
	"os"
	"testing"
	"time"
)

func getRedisConfigStr() string {
	//make sure you have this env setup
	return os.Getenv(`REDIS_CONFIG_STR`)
}

func TestDlock(t *testing.T) {
	cfg := map[string]string{
		`redis_config_string`: getRedisConfigStr(),
	}
	dlock, err := NewDlock(cfg)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer dlock.Close()
	key := `test-a-unique-key`
	m := dlock.NewMutex(context.Background(), key)
	if err := m.Lock(); err != nil {
		t.Fatal(err.Error())
	}

	t.Log(`do second lock`)
	if err := m.Lock(); err != nil {
		t.Log(`second attemp shall not success before the first lock do Unlock`, err.Error())
	}
	time.Sleep(1 * time.Second)
	t.Log(`unlock first  `)
	if err := m.Unlock(); err != nil {
		t.Fatal(err.Error())
	}
	//below is testing if a lock can repeatly used safely. !!!not recommended
	t.Log(`lock third  `)
	if err := m.Lock(); err != nil {
		t.Fatal(`failed on released lock acquiring`, err.Error())
	}
	time.Sleep(2 * time.Second)
	t.Log(`unlock third`)
	if err := m.Unlock(); err != nil {
		t.Fatal(err.Error())
	}

	t.Log(`lock fourth`)
	if err := m.Lock(); err != nil {
		t.Fatal(`failed on released lock acquiring`, err.Error())
	}

	time.Sleep(3 * time.Second)
	t.Log(`unlock fourth  `)
	if err := m.Unlock(); err != nil {
		t.Fatal(err.Error())
	}

	return

}

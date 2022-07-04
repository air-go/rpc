package logger

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

// LogId is a unique ID identifying a log record. It must be exactly 12 bytes
// long.
//
// Reference:http://www.mongodb.org/display/DOCS/Object+IDs
type ObjectID [12]byte

var (
	// objectIDCounter is atomically incremented when generating a new ObjectID
	// using NewObjectID() function. It's used as a counter part of an id.
	objectIDCounter uint64 = 0

	// machineID stores machine id generated once and used in subsequent calls
	// to NewObjectID function.
	machineID = initMachineID()
)

// readMachineId generates machine id and puts it into the machineID global
// variable. If this function fails to get the hostname, it will cause
// a runtime error.
func initMachineID() []byte {
	var sum [3]byte
	id := sum[:]
	hostname, err1 := os.Hostname()
	if err1 != nil {
		_, err2 := io.ReadFull(rand.Reader, id)
		if err2 != nil {
			panic(fmt.Errorf("cannot get hostname: %v; %v", err1, err2))
		}
		return id
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(id, hw.Sum(nil))
	return id
}

// NewObjectID returns a new unique ObjectID.
func NewObjectID() ObjectID {
	var b [12]byte
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b[:], uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	b[4] = machineID[0]
	b[5] = machineID[1]
	b[6] = machineID[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	pid := os.Getpid()
	b[7] = byte(pid >> 8)
	b[8] = byte(pid)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint64(&objectIDCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	return b
}

// NewObjectIDWithTime returns a dummy ObjectID with the timestamp part filled
// with the provided number of seconds from epoch UTC, and all other parts
// filled with zeroes. It's not safe to insert a document with an id generated
// by this method, it is useful only for queries to find documents with ids
// generated before or after the specified timestamp.
func NewObjectIDWithTime(t time.Time) ObjectID {
	var b [12]byte
	binary.BigEndian.PutUint32(b[:4], uint32(t.Unix()))
	return b
}

func NewObjectIDWithHexString(s string) (o ObjectID, err error) {
	d, err := hex.DecodeString(s)
	if err != nil || len(d) != 12 {
		err = errors.New("string length must 12")
		return
	}
	copy(o[:], d[:12])
	return
}

// Hex returns a hex representation of the ObjectID.
func (id ObjectID) Hex() string {
	return hex.EncodeToString(id[:])
}

// Time returns the timestamp part of the id.
// It's a runtime error to call this method with an invalid id.
func (id ObjectID) Time() time.Time {
	// First 4 bytes of ObjectID is 32-bit big-endian seconds from epoch.
	secs := int64(binary.BigEndian.Uint32(id[:4]))
	return time.Unix(secs, 0)
}

// Machine returns the 3-byte machine id part of the id.
// It's a runtime error to call this method with an invalid id.
func (id ObjectID) Machine() []byte {
	return id[4:7]
}

// Pid returns the process id part of the id.
// It's a runtime error to call this method with an invalid id.
func (id ObjectID) Pid() uint16 {
	return binary.BigEndian.Uint16(id[7:9])
}

// Counter returns the incrementing value part of the id.
// It's a runtime error to call this method with an invalid id.
func (id ObjectID) Counter() int32 {
	b := id[9:12]
	// Counter is stored as big-endian 3-byte value
	return int32(uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2]))
}

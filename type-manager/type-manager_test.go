package typemanager_test

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	"github.com/touee/nyn/data-packer"

	typemanager "github.com/touee/nyn/type-manager"
)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))

}

func tempFilePath() string {
	return path.Join(os.TempDir(), fmt.Sprintf("test_%d.s3db", rand.Int()))
}

func TestStorage(t *testing.T) {
	var err error
	var fileName = tempFilePath()
	defer func() {
		os.Remove(fileName)
	}()

	var m *typemanager.TypeManager
	m, err = typemanager.OpenTypeManager(fileName, datapacker.DefaultPacker)
	if err != nil {
		t.Fatal(err)
	}

	type T1 struct {
		X int
	}

	err = m.Register(T1{})
	if err != nil {
		t.Fatal(err)
	}

	err = m.Close()
	if err != nil {
		t.Fatal(err)
	}

	// 重新打开该 TypeManager
	if m, err = typemanager.OpenTypeManager(fileName, datapacker.DefaultPacker); err != nil {
		t.Fatal(err)
	}

	{ //< 通过注册一个与之前注册类型的类型名相同但签名不同的类型, 判断返回的错误是否是签名类型不同, 来判断注册的类型是否被成功存储
		type T1 struct{}
		err = m.Register(T1{})
		if sErr, ok := err.(typemanager.SignatureNotMatchError); ok {
			t.Log(sErr.Error())
		} else {
			t.Fatal(err)
		}
	}
}

func TestPackAndUnpack(t *testing.T) {
	var err error
	var fileName = tempFilePath()
	defer func() {
		os.Remove(fileName)
	}()

	var m *typemanager.TypeManager
	m, err = typemanager.OpenTypeManager(fileName, datapacker.DefaultPacker)
	if err != nil {
		t.Fatal(err)
	}

	type T1 struct {
		X int
	}

	err = m.Register(T1{})
	if err != nil {
		t.Fatal(err)
	}

	var task = T1{X: 123}

	var packed []byte
	packed, err = m.Pack(task)
	if err != nil {
		t.Fatal(err)
	}

	var unpacked interface{}
	unpacked, err = m.Unpack(typemanager.GetTypeName(task), packed)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%v", task) != fmt.Sprintf("%v", unpacked) {
		t.Fatal(task, "≠", unpacked)
	}

	err = m.Close()
	if err != nil {
		t.Fatal(err)
	}

	// 重新打开该 TypeManager
	if m, err = typemanager.OpenTypeManager(fileName, datapacker.DefaultPacker); err != nil {
		t.Fatal(err)
	}

	err = m.Register(T1{})
	if err != nil {
		t.Fatal(err)
	}

	unpacked, err = m.Unpack(typemanager.GetTypeName(task), packed)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%v", task) != fmt.Sprintf("%v", unpacked) {
		t.Fatal(task, "≠", unpacked)
	}

}

type FakePacker struct{}

func (p FakePacker) Formats() (string, string)                                { return "a", "b" }
func (p FakePacker) Pack(task interface{}) (packedData []byte, err error)     { panic("") }
func (p FakePacker) Unpack(data []byte, task interface{}) (err error)         { panic("") }
func (p FakePacker) Signature(task interface{}) (signature string, err error) { panic("") }

func TestErrors(t *testing.T) {
	var err error
	var fileName = tempFilePath()
	defer func() {
		os.Remove(fileName)
	}()

	var m *typemanager.TypeManager
	m, err = typemanager.OpenTypeManager(fileName, datapacker.DefaultPacker)
	if err != nil {
		t.Fatal(err)
	}

	// 测试打包/解包未注册的任务
	_, err = m.Pack("")
	if rErr, ok := err.(typemanager.TypeNotRegisteredError); ok {
		t.Log(rErr.Error())
	} else {
		t.Fatal(err)
	}

	_, err = m.Unpack("", nil)
	if rErr, ok := err.(typemanager.TypeNotRegisteredError); ok {
		t.Log(rErr.Error())
	} else {
		t.Fatal(err)
	}

	// 测试不匹配的 Packer
	err = m.Close()
	if err != nil {
		t.Fatal(err)
	}

	m, err = typemanager.OpenTypeManager(fileName, FakePacker{})
	if fErr, ok := err.(typemanager.PackFormatOrSignatureFormatNotMatchError); ok {
		t.Log(fErr.Error())
	} else {
		t.Fatal(err)
	}

}

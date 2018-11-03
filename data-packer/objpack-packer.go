package datapacker

import objpack "github.com/touee/go-objpack"

// ObjpackPacker 是通过包装 "github.com/touee/nyn/objpack" 包中的功能实现的 Packer
type ObjpackPacker struct{}

// Formats 返回 "objpack", "objpack-signature"
func (packer ObjpackPacker) Formats() (packFormat string, signatureFormat string) {
	return "objpack", "objpack-signature"
}

// Pack 使用 objpack 提供的 Pack 函数打包任务
func (packer ObjpackPacker) Pack(task interface{}) (packedData []byte, err error) {
	return objpack.Pack(task)
}

// Unpack 使用 objpack 提供的 Unpack 函数取出被打包的任务
func (packer ObjpackPacker) Unpack(data []byte, task interface{}) (err error) {
	return objpack.Unpack(data, task)
}

// Signature 使用 objpack 提供的 Signature 生成任务类型的签名
func (packer ObjpackPacker) Signature(task interface{}) (signature string, err error) {
	return objpack.MakeTypeSignature(task)
}

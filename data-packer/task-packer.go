package datapacker

// Packer 是用于 pack/unpack 任务的接口
type Packer interface {
	// Formats 用于标明任务被打包的成的格式, 以及对类型生成签名的格式
	Formats() (packFormat string, signatureFormat string)

	// Pack 用于打包任务
	Pack(task interface{}) (packedData []byte, err error)
	// Unpack 用于从打包的数据取出任务
	Unpack(data []byte, task interface{}) (err error)
	// Signature 用于声明任务类型的签名, 以防错误使用
	Signature(task interface{}) (signature string, err error)
}

// DefaultPacker 是默认的 packer, 即 ObjpackPacker{}
var DefaultPacker = ObjpackPacker{}

package typemanager

import "fmt"

// SignatureNotMatchError 是当尝试注册任务类型, 但其签名与数据库中所存的任务类型的签名不符时, 返回的错误
type SignatureNotMatchError struct {
	Name                          string
	wantedSignature, gotSignature string
}

func (e SignatureNotMatchError) Error() string {
	return fmt.Sprintf(`nyn: Task type signature not match: type name: %s, wanted "%s", got: "%s"`,
		e.Name, e.wantedSignature, e.gotSignature)
}

// TypeNotRegisteredError 是尝试打包/解包任务, 但任务类型并没有被注册时返回的错误
type TypeNotRegisteredError struct {
	Name string
}

func (e TypeNotRegisteredError) Error() string {
	return fmt.Sprintf("nyn: Task type not registered: %s", e.Name)
}

// PackFormatOrSignatureFormatNotMatchError 是传入 TaskTypeManager 的 Packer 所给输出的格式与预期格式不匹配时返回的错误
type PackFormatOrSignatureFormatNotMatchError struct {
	usedPackFormat, usedSignatureFormat         string
	expectedPackFormat, expectedSignatureFormat string
}

func (e PackFormatOrSignatureFormatNotMatchError) Error() string {
	return fmt.Sprintf("nyn: pack format or signature format not match. expected pack format: %s, used pack format: %s; expected signature format: %s, used signature format: %s", e.expectedPackFormat, e.usedPackFormat, e.expectedSignatureFormat, e.usedSignatureFormat)
}

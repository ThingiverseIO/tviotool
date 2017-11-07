package thing

type Compiler interface {
	Compile(target string) error 
}

package gentest

//go:generate sortnetgen -o primitive_gen.go -fwd -rev -size 2-16,24,32,48,64 string int
//go:generate sortnetgen -o custom_gen.go -fwd -rev -size 2-16,24,32,48,64 -less CustomCASLess -greater CustomCASGreater Custom

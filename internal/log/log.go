package log

import "fmt"

func Debug(obj ...any) {
	obj = append([]any{"[D] "}, obj...)
	fmt.Println(obj...)
}

func Debugf(f string, obj ...any) {
	Debug(fmt.Sprintf(f, obj...))
}

func Info(obj ...any) {
	obj = append([]any{"[I] "}, obj...)
	fmt.Println(obj...)
}

func Infof(f string, obj ...any) {
	Info(fmt.Sprintf(f, obj...))
}

func Warn(obj ...any) {
	obj = append([]any{"[W] "}, obj...)
	fmt.Println(obj...)
}

func Warnf(f string, obj ...any) {
	Warn(fmt.Sprintf(f, obj...))
}

func Err(obj ...any) {
	obj = append([]any{"[E] "}, obj...)
	fmt.Println(obj...)
}

func Errf(f string, obj ...any) {
	Err(fmt.Sprintf(f, obj...))
}

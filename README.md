# gosif

`gosif` is a tool that helps you to create a simple CLI for your Go app.

## What is gosif

`gosif` is a simple and lite tool that generates CLI for Go apps. More precisely, it generates all the necessary functions to pass arguments to your app via a CLI and to check their correctness.

## Why gosif

Go is a great language. Having a codebase written entirely in Go it's a shame to have to use `bash`, or `python`, or *whatnot* for scripting. Initially `gosif` was designed to allow developers to write scripts without wasting time for typing in a lot of bolierplate for parsing, checking and passing arguments to the scripts functions.

At the same time, we wanted to create a very light tool, without tons of dependencies to maintain. As a result `gosif` (as well as the code that it generates) depends only on the standard library.

### When you should not use gosif

`gosif` is an opinionated tool that does not allow to build a rich CLI  out of the box. We wanted to keep `gosif` simple: there is no configuration whatsoever. That being said, you can tweak the generated files as you want, but the interface that `gosif` generates is rather succinct.

## Contents

- [Quick start](#quick-start)
- [How to use gosif](#how-to-use-gosif)
- [How gosif processes your application](#how-gosif-processes-your-application)
- [Generated help messages](#generated-help-messages)
- [Argument-types](#argument-types)
	- [String](#string)
	- [Byte](#byte)
	- [Rune](#rune)
	- [Bool](#bool)
	- [Signed integers](#signed-integers)
	- [Unsigned integers](#unsigned-integers)
	- [Floating-point numbers](#floating-point-numbers)
	- [Complex numbers](#complex-numbers)
	- [Arguments of the error type](#arguments-of-the-error-type)
	- [Slices and arrays](#slices-and-arrays)
	- [Pointers](#pointers)
	- [Slices and pointers combination](#slices-and-pointers-combination)
	- [Arrays and pointers combination](#arrays-and-pointers-combination)
- [License](#license)

## Quick start

Let's see `gosif` in action. Here is a simple, "Hello, World!" example to outline what `gosif` does. Let's say you want to write an app that joins the passed arguments into a string and optionally converts the result to uppercase. Such a function may look like:

```go
func PrintStringMaybeUpper(parts []string, upper bool) {
	str := strings.Join(parts, " ")
	if upper {
		str = strings.ToUpper(str)
	}
	fmt.Println(str)
}
```

To generate an interface for it we pass the directory that contains this file to `gosif`:

```bash
gosif maybe_upper_string
```

`gosif` generates a file `main.gen.go` with a main function that parses the CLI arguments for your application and passes it to `PrintStringMaybeUpper`. Running the app prints:

```bash
go run . PrintStringMaybeUpper --parts Hello gosif ! --upper
> HELLO GOSIF !
```

`gosif` also generates validators and help messages for your functions. So, if you forget to pass the expected argument `--parts` you get a reminder:

```bash
go run . PrintStringMaybeUpper --upper
> [ERR]: a required flag "-parts" was not passed
> Function PrintStringMaybeUpper
> 	Required options:
>		--parts     []string
>	Available options:
>		--parts     []string
>		--upper     bool
```

The code for this example is in [examples/readme/maybe_upper_string/myscript.go](examples/readme/maybe_upper_string/myscript.go)

## How to use gosif

To use `gosif` for your application you need to:

1. install `gosif`:

```bash
go get github.com/SergeyShpak/gosif
```

2. pass the name of the folder that contains your application to `gosif`:

```bash
gosif my-app/
```

3. build and run your application or use `go run`:

```bash
cd my-app/ && go build . && ./my-app <args>
# or
cd my-app && go run . <args>
```

If you run `gosif` on the directory that already contains a file `main.gen.go`, `gosif` will scan this file. If it contains only functions `main()`, `gosif()` and functions prefixed with `gosif_`, `gosif` rewrites it. If `gosif` finds other functions inside the file it shows the error message and cancels the code generation.

## How gosif processes your application

`gosif` generates CLI for executables. It scans through the package `main`, finds all the exportable functions and tries to generate interfaces for them. It skips functions that it cannot process.

A function that `gosif` can process:
1. has arguments of types that are listed in the [Argument types](#argument-types) section only
2. does not return anything
3. is exportable (its name starts with a capital letter)
4. are located in the `main` package

If the `main` package does not contain the `main()` function yet, `gosif` generates it. Otherwise, `gosif` generates a function `gosif()` that should be manually added to `main()`.

## Generated help messages

`gosif` generates help messages for your application functions that indicate names of the available functions, names of the function flags and types of the expected arguments.

To see the list of executable functions of your app run:

```bash
go run . help
> The following functions are available:
> ...
```

This message is also shown when trying to run a function that does not exist:

```bash
go run . InexistentFunc
> [ERR]: unknown script InexistentFunc
> The following functions are available:
> ...
```

To see the help message of a function, run the function with the argument `help`:

```bash
go run . MyFunc help
> Function MyFunc:
> 	Required options:
>	...
>	Available options:
>	...
```

This message is also shown when passing a bad argument to the function:

```bash
go run . OnlyInts --n 3.14
> [ERR]: cast failed: failed to cast 12.34 to int: strconv.ParseInt: parsing "3.14": invalid syntax
> Function OnlyInts:
> ...
```

## Argument types

`gosif` can generate interfaces for functions with arguments of the following types:

- string
- byte
- rune
- bool
- int
- int8
- int16
- int32
- int64
- uint
- uint8
- uint16
- uint32
- uint64
- float32
- float64
- complex64
- complex128
- error

You can find the code from this section in [examples/readme/types_demo/myscript.go](examples/readme/types_demo/myscript.go)

### String

We use this function in the example:

```go
func StringType(s string) {
	fmt.Println(s)
}
```

Passing a string argument results in:

```bash
go run . StringType --s hello!
> hello!
```

`gosif` functions pass the string argument as is, ignoring escape sequences.

```bash
go run . StringType --s 'Hello!\n'
> Hello!\n
```

The only exception to this are double quotes. A double-quoted chain of characters is treated as a single argument.

```bash
go run . StringType --s "Hello, gosif!"
> Hello, gosif!
```

### Byte

We use this function in the example:

```go
func ByteType(b byte) {
	fmt.Println(b)
}
```

A byte argument should be passed as a decimal, 8-bit unsigned integer number:

```bash
go run . ByteType --b 42
> 42
```

### Rune

We use this function in the example:

```go
func RuneType(r rune) {
	fmt.Println(string(r))
}
```

A rune argument should be passed as a utf8 character:

```bash
go run . RuneType --r 😀
> 😀
```

### Bool

We use this function in the example:

```go
func BoolType(bt bool, bf bool) {
	fmt.Println(bt, bf)
}
```

There are three formats for boolean arguments:

1. You can pass the boolean flag without arguments for a `true` value, or omit it for a `false` value:

```bash
go run . BoolType --bt
> true false
```

2. You can pass a single letter `t` for a `true` value, or `f` for a `false` value:

```bash
go run . BoolType --bt t --bf f
> true false
```

3. You can pass a word `true` for a `true` value, or `false` for a `false` value:

```bash
go run . BoolType --bt true --bf false
> true false
```

The passed arguments are case-insensitive:

```bash
go run . BoolType --bt T --bf False
> true false
```

Using arguments that are pointers to bools is slightly different, see [*pointers*](#pointers-bool-pointer) section for details.

### Signed integers

We use this function in the example:

```go
func SignedIntType(n int, n8 int8, n16 int16, n32 int32, n64 int64) {
	fmt.Printf("n: %d, n8: %d, n16: %d, n32: %d, n64: %d\n", n, n8, n16, n32, n64)
}
```

A passed string argument is converted to a signed integer by `strconv.ParseInt` (the base of the integer is implied by the passed string prefix), see its documentation to learn about accepted string formats:

```bash
go run . SignedIntType --n 0b101010 --n8 -8 --n16 +16 --n32 040 --n64 -0x40
> n: 42, n8: -8, n16: 16, n32: 32, n64: -64
```

### Unsigned integers

We use this function in the example:

```go
func UnsignedIntType(n uint, n8 uint8, n16 uint16, n32 uint32, n64 uint64) {
	fmt.Printf("n: %d, n8: %d, n16: %d, n32: %d, n64: %d\n", n, n8, n16, n32, n64)
}
```

A passed string argument is converted to an unsigned integer by `strconv.ParseUint` (the base of the integer is implied by the passed string prefix), see its documentation to learn about accepted string formats:

```bash
go run . UnsignedIntType --n 42 --n8 0b1000 --n16 020 --n32 0o40 --n64 0x40
> n: 42, n8: 8, n16: 16, n32: 32, n64: 64
```

### Floating-point numbers

We use this function in the example:

```go
func FloatType(f32 float32, f64 float64) {
	fmt.Printf("f32: %.3f, f64: %.3f\n")
}
```

A passed string argument is converted to a floating-point number by `strconv.ParseFloat` (the base of the integer is implied by the passed string prefix), see its documentation to learn about accepted string formats:

```bash
go run . FloatType --f32 3.14159 --f64 -6.02E+23
> f32: 3.142, f64: -601999999999999995805696.000
```

### Complex numbers

We use this function in the example:

```go
func ComplexType(c64 complex64, c128 complex128) {
	fmt.Printf("c64: (%.3f, %.3fi), c128: (%.3f, %.3fi)\n", real(c64), imag(c64), real(c128), imag(c128)
}
```

There are two valid ways to represent a complex number:

1. If one of the complex number parts (real or imaginary) is zero, then you can pass the other part directly as the function argument. The imaginary part should have the suffix `i`.

```bash
go run . ComplexType --c64 42 --c128 -42i
> c64: (42.000, 0.000i), c128: (0.000, -42.000i)
```

2. A standard way of representing a complex number is putting the real and imaginary parts between parentheses and separating them with a comma. Parts can be placed in an arbitrary order, but the imaginary part should have the suffix `i`. Please note that some shells only accept quoted arguments with parentheses.

```bash
go run . ComplexType --c64 "(3.14159, 0i)" --c128 "(-i, 42.123)"
> c64: (3.142, 0.000i), c128: (42.123, -1.000i)
```

We use `strconv.ParseFloat` to parse the real and imaginary parts, so, to represent them, you can use any float-point numbers syntax valid for this function (see its documentation for details):

```bash
go run . ComplexType --c64 "(+Inf, NaNi)" --c128 -6.02E+23i
> c64: (+Inf, NaNi), c128: (0.000, -601999999999999995805696.000i)
```

### Arguments of the error type

We use this function in the example:

```go
func ErrorType(e error) {
	fmt.Println(e)
}
```

Passing a string as an error argument makes the string the error message:

```bash
go run . ErrorType --e "this is an error message"
> this is an error message
```

### Slices and arrays

You can use slices and arrays of the [available types](#available-arguments-types) in your functions definitions:

```go
func MySliceFunc(nums []int) { fmt.Println(nums) }
func MyArrFunc(nums [3]int) { fmt.Println(nums) }
```

In case of a slice, `gosif` functions treat the passed arguments separated by spaces as the elements of the slice.

```bash
go run . MySliceFunc --nums 1 2 3
> [1 2 3]
go run . MySliceFunc --nums
> []
```

The same is valid for arrays arguments:

```bash
go run . MyArrFunc --nums 1 2 3
> [1 2 3]
```

However, `gosif` treats slices and arrays semantically different: `gosif` functions check that a correct number of arguments were passed for an array argument:

```bash
go run . MyArrFunc --nums 1 2
> [ERR]: flag --nums: expected 3 arguments, but got 2
```

As multi-dimensional slices and arrays are currently not supported, `gosif` skips functions with arguments of such a type. Running `gosif` on 

```go
func MyMultiSliceFunc(nums [][][]int) { fmt.Println(nums) }
func MyMultiArrFunc(nums [2][3]int)   { fmt.Println(nums) }
```

results is a warning:

```bash
[WARN]: skipping the function MyMultiSliceFunc in examples/readme/slices_and_arrays/myscript.go: failed to parse the parameter "nums": multidimensional parameters are not yet supported
...
[WARN]: skipping the function MyMultiArrFunc in examples/readme/slices_and_arrays/myscript.go: failed to parse the parameter "nums": multidimensional parameters are not yet supported
```

You can find the code used in this section in [examples/readme/slices_and_arrays/myscript.go](examples/readme/slices_and_arrays/myscript.go)

### Pointers

You can use pointers of the [available types](#available-arguments-types) in your functions definitions:

```go
func MyPointerFunc(num *int) {
	if num == nil {
		fmt.Println("nil")
	} else {
		fmt.Println(*num)
	}
}
```

`gosif` functions treat pointer arguments as optional:

```bash
go run . MyPointerFunc --num 42
> 42
go run . MyPointerFunc
> nil
```

Passing no arguments to the optional flag, however, results in an error:

```bash
go run . MyPointerFunc --num
> [ERR]: could not get the argument passed to the flag "--num": no arguments passed
> ...
```

`gosif` correctly treats multiple levels of indirection (pointers to pointers to pointers...):

```go
func MyManyPointersFunc(num ***int) {
	if num == nil {
		fmt.Println("nil")
	} else {
		fmt.Println(***num)
	}
}
```

As `gosif` does not treat the passed arguments differently, we may check only if the first pointer is nil in `MyManyPointersFunc`. With `gosif` you cannot pass a pointer to a nil pointer.

Running the example with multiple pointers gives the same result as for the one with the single pointer:

```bash
go run . MyManyPointersFunc --num 42
> 42
go run . MyManyPointersFunc
> nil
go run . MyManyPointersFunc --num
> [ERR]: could not get the argument passed to the flag "--num": no arguments passed
> ...
```

<a name="pointers-bool-pointer"></a>Working with pointers to booleans is different then working with direct boolean values: you may no longer omit the boolean flag to get a `false` value, you need to specify that the argument is false explicitly. Running the function

```go
func BoolPointerFunc(b *bool) {
	if b == nil {
		fmt.Println("nil")
	} else {
		fmt.Println(*b)
	}
}
```

outputs

```bash
go run . BoolPointerFunc --b
> true
go run . BoolPointerFunc
> nil
go run . BoolPointerFunc --b f
> false
```

You can find the code used in this section in [examples/readme/pointers/myscript.go](examples/readme/pointers/myscript.go)

### Slices and pointers combination

`gosif` treats arguments that are pointer to slices of some type as optional. Arguments that are slices of pointers are treated the same way as slices of direct values, i.e. with `gosif` it is not possible to pass a slice that contains `nil`s to the function.

Running the function

```go
func MySlicePointerFunc(ps *[]int, sp []*int, psp *[]*int) {
	if ps == nil {
		fmt.Println("ps: nil")
	} else {
		fmt.Println("ps: ", *ps)
	}
	spElems := make([]int, len(sp))
	for i, el := range sp {
		spElems[i] = *el
	}
	fmt.Println("sp: ", spElems)
	if psp == nil {
		fmt.Println("psp: nil")
	} else {
		pspElems := make([]int, len(*psp))
		for i, el := range *psp {
			pspElems[i] = *el
		}
		fmt.Println("psp: ", pspElems)
	}
}
```

outputs

```bash
go run . MySlicePointerFunc --ps 1 2 3 --sp 1 2 3 --psp 1 2 3
> ps:  [1 2 3]
> sp:  [1 2 3]
> psp:  [1 2 3]
```

Multiple levels of indirection (pointers to pointers) are allowed, you can find more information about them in the section [*pointers*]($pointers).

You can find the code used in this section in [examples/readme/slices_arrays_pointers/slices_pointers.go](examples/readme/slices_arrays_pointers/slices_pointers.go)

### Arrays and pointers combination

`gosif` treats arguments that are pointer to arrays of some type as optional. Arguments that are arrays of pointers are treated differently from [slices of pointers](#slices-and-pointers-combination): elements of such arrays are optional, so the missing elements are set to nil:

Running the function

```go
func MyArrPointerFunc(pa *[3]int, ap [3]*int, pap *[3]*int) {
	if pa == nil {
		fmt.Println("pa: nil")
	} else {
		fmt.Println("pa: ", *pa)
	}
	apElems := make([]string, len(ap))
	for i, el := range ap {
		if el == nil {
			apElems[i] = "nil"
		} else {
			apElems[i] = fmt.Sprintf("%d", *el)
		}
	}
	fmt.Println("ap: ", apElems)
	if pap == nil {
		fmt.Println("pap: nil")
	} else {
		papElems := make([]string, len(*pap))
		for i, el := range *pap {
			if el == nil {
				papElems[i] = "nil"
			} else {
				papElems[i] = fmt.Sprintf("%d", *el)
			}
		}
		fmt.Println("pap: ", papElems)
	}
}
```

outputs

```bash
go run . MyArrPointerFunc --ap 1 2 --pap
> pa: nil
> ap:  [1 2 nil]
> pap:  [nil nil nil]
```

Multiple levels of indirection (pointers to pointers) are allowed, you can find more information about them in the section [*pointers*]($pointers).

You can find the code used in this section in [examples/readme/slices_arrays_pointers/arrays_pointers.go](examples/readme/slices_arrays_pointers/arrays_pointers.go)

## License

See the [LICENSE](LICENSE.md) file for license rights and limitations (MIT).

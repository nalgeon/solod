package clang

import (
	"go/ast"
	"go/types"
)

// reservedAnyScope contains C identifiers that conflict in every scope.
var reservedAnyScope = map[string]bool{
	// C keywords that are valid Go identifiers.
	"auto": true, "char": true, "do": true, "double": true, "enum": true,
	"extern": true, "float": true, "inline": true, "int": true, "long": true,
	"register": true, "restrict": true, "short": true, "signed": true,
	"sizeof": true, "static": true, "typedef": true, "union": true,
	"unsigned": true, "void": true, "volatile": true, "while": true,
	// C23 keywords and predeclared Go identifiers that map to C keywords/macros.
	"bool": true, "true": true, "false": true, "nullptr": true,
	"alignas": true, "alignof": true, "constexpr": true, "typeof": true,
	"thread_local": true, "static_assert": true,
	// builtin.h macros.
	"alloca": true, "assert": true, "offsetof": true, "NULL": true,
	"memcmp": true, "memcpy": true, "memmove": true, "memset": true,
	// stdio.h macros.
	"EOF": true, "stdin": true, "stdout": true, "stderr": true,
	// errno.h macros.
	"errno": true,
}

// reservedFileScope contains C identifiers that only conflict in file scope.
var reservedFileScope = map[string]bool{
	// stdio.h, C standard.
	"clearerr": true, "fclose": true, "feof": true, "ferror": true,
	"fflush": true, "fgetc": true, "fgetpos": true, "fgets": true,
	"fopen": true, "fprintf": true, "fputc": true, "fputs": true,
	"fread": true, "freopen": true, "fscanf": true, "fseek": true,
	"fsetpos": true, "ftell": true, "fwrite": true, "getc": true,
	"getchar": true, "gets": true, "perror": true, "printf": true,
	"putc": true, "putchar": true, "puts": true, "remove": true,
	"rename": true, "rewind": true, "scanf": true, "setbuf": true,
	"setvbuf": true, "snprintf": true, "sprintf": true, "sscanf": true,
	"tmpfile": true, "tmpnam": true, "ungetc": true, "vfprintf": true,
	"vfscanf": true, "vprintf": true, "vscanf": true, "vsnprintf": true,
	"vsprintf": true, "vsscanf": true,
	// stdio.h, POSIX and BSD.
	"asprintf": true, "dprintf": true, "fdopen": true, "fileno": true,
	"flockfile": true, "fmemopen": true, "fseeko": true, "ftello": true,
	"ftrylockfile": true, "funlockfile": true, "getdelim": true,
	"getline": true, "getw": true, "pclose": true, "popen": true,
	"putw": true, "setbuffer": true, "setlinebuf": true, "tempnam": true,
	"vasprintf": true, "vdprintf": true,
	// stdlib.h, C standard.
	"abort": true, "abs": true, "aligned_alloc": true, "at_quick_exit": true,
	"atexit": true, "atof": true, "atoi": true, "atol": true, "atoll": true,
	"bsearch": true, "calloc": true, "div": true, "exit": true, "free": true,
	"getenv": true, "labs": true, "ldiv": true, "llabs": true, "lldiv": true,
	"malloc": true, "mblen": true, "mbstowcs": true, "mbtowc": true,
	"qsort": true, "quick_exit": true, "rand": true, "realloc": true,
	"srand": true, "strtod": true, "strtof": true, "strtol": true,
	"strtold": true, "strtoll": true, "strtoul": true, "strtoull": true,
	"system": true, "wcstombs": true, "wctomb": true,
	// stdlib.h, POSIX and BSD.
	"a64l": true, "arc4random": true, "daemon": true, "drand48": true,
	"ecvt": true, "erand48": true, "fcvt": true, "gcvt": true,
	"getloadavg": true, "getprogname": true, "getsubopt": true,
	"grantpt": true, "heapsort": true, "initstate": true, "jrand48": true,
	"l64a": true, "lcong48": true, "lrand48": true, "mergesort": true,
	"mkdtemp": true, "mkstemp": true, "mktemp": true, "mrand48": true,
	"nrand48": true, "ptsname": true, "putenv": true, "radixsort": true,
	"random": true, "realpath": true, "reallocf": true, "rpmatch": true,
	"seed48": true, "setenv": true, "setprogname": true, "setstate": true,
	"srand48": true, "srandom": true, "unlockpt": true, "unsetenv": true,
	"valloc": true,
	// stdlib.h drags in these declarations on Darwin.
	"getpriority": true, "getrlimit": true, "getrusage": true,
	"setpriority": true, "setrlimit": true, "signal": true, "wait": true,
	"wait3": true, "wait4": true, "waitid": true, "waitpid": true,
	// string.h, C standard. The memory functions are in reservedAnyScope,
	// because builtin.h makes them macros in a freestanding build.
	"memchr": true, "strcat": true, "strchr": true, "strcmp": true,
	"strcoll": true, "strcpy": true, "strcspn": true, "strerror": true,
	"strlen": true, "strncat": true, "strncmp": true, "strncpy": true,
	"strpbrk": true, "strrchr": true, "strspn": true, "strstr": true,
	"strtok": true, "strxfrm": true,
	// string.h, POSIX and BSD.
	"bcmp": true, "bcopy": true, "bzero": true, "ffs": true, "fls": true,
	"index": true, "memccpy": true, "memmem": true, "mempcpy": true,
	"rindex": true, "stpcpy": true, "stpncpy": true, "strcasecmp": true,
	"strcasestr": true, "strchrnul": true, "strdup": true, "strlcat": true,
	"strlcpy": true, "strncasecmp": true, "strndup": true, "strnlen": true,
	"strnstr": true, "strsep": true, "strsignal": true, "strverscmp": true,
	"swab": true,
	// inttypes.h.
	"imaxabs": true, "imaxdiv": true, "strtoimax": true, "strtoumax": true,
	"wcstoimax": true, "wcstoumax": true,
}

// reservedSo contains identifiers that the generator emits itself.
var reservedSo = map[string]bool{
	// The method receiver parameter and the interface data pointer.
	"self": true,
}

// resolveReservedNames renames or rejects the identifiers with a reserved
// C name. It also rejects the identifiers with a [reservedSo] name.
func (g *Generator) resolveReservedNames() {
	for _, file := range g.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if ident, ok := n.(*ast.Ident); ok {
				g.resolveReservedName(ident)
			}
			return true
		})
	}
}

// resolveReservedName handles one identifier.
func (g *Generator) resolveReservedName(ident *ast.Ident) {
	if reservedSo[ident.Name] {
		// Report the name once, at the declaration.
		if _, isDef := g.types.Defs[ident]; isDef {
			g.fail(ident, "identifier %q is reserved by the compiler; rename it", ident.Name)
		}
		return
	}
	obj, isDef := g.types.Defs[ident]
	if !isDef {
		obj = g.types.Uses[ident]
	}
	if obj == nil {
		// An identifier that represents no object, like the package clause
		// name or a blank identifier. It never becomes a C identifier.
		return
	}
	if !g.makesCName(obj) {
		return
	}
	name := g.baseObjName(obj)
	if !isReserved(name, g.atPkgScope(obj)) {
		return
	}
	if !g.canRename(obj) {
		if isDef {
			g.fail(ident, "identifier %q is a reserved C name; rename it", name)
		}
		return
	}
	mangled := name + "_"
	if isDef && obj.Parent().Lookup(mangled) != nil {
		// Only the scope of the renamed object matters. A C redefinition only
		// happens within a single block, and go/types scopes match C blocks
		// one-to-one. A name in an outer or inner scope is a valid C shadow.
		g.fail(ident, "mangled name %q for reserved C name %q conflicts with an existing identifier; rename one of them", mangled, name)
	}
	g.renames[obj] = mangled
	if g.isLocalValue(obj) {
		// Some emit sites take a local name from its identifier instead of
		// its object, so rename the identifier too.
		ident.Name = mangled
	}
}

// makesCName reports whether the generator makes the C name of obj
// from the Go source of this package.
func (g *Generator) makesCName(obj types.Object) bool {
	if obj.Pkg() != g.pkg.Types {
		// A predeclared Go identifier, or an object of another package.
		return false
	}
	if _, isImport := obj.(*types.PkgName); isImport {
		// An import alias never becomes a C identifier. The generator emits
		// the name of the imported package.
		return false
	}
	if _, isExtern := g.getExtern(obj); isExtern {
		// The C name of an extern object comes from the C source.
		return false
	}
	return true
}

// canRename reports whether the generator can change the C name of obj.
func (g *Generator) canRename(obj types.Object) bool {
	if g.isLocalValue(obj) {
		// Never leave the .c file, safe to rename.
		return true
	}
	// Unexported exported package-level names never leave the .c file.
	// Exported package-level names are renamed when emitted.
	return g.atPkgScope(obj)
}

// isLocalValue reports whether obj is a function-local
// variable, parameter, or constant.
func (g *Generator) isLocalValue(obj types.Object) bool {
	switch obj.(type) {
	case *types.Var, *types.Const:
	default:
		return false
	}
	// A struct field has no parent scope.
	parent := obj.Parent()
	return parent != nil && parent != g.pkg.Types.Scope()
}

// isReserved reports whether C already uses name.
func isReserved(name string, fileScope bool) bool {
	return reservedAnyScope[name] || (fileScope && reservedFileScope[name])
}

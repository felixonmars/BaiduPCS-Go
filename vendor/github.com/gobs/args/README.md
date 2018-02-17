args
====

command line argument parser

Given a string (the "command line") it splits into words separated by white spaces, respecting quotes (single and double) and the escape character (backslash)

## Installation

    $ go get github.com/gobs/args

## Documentation
http://godoc.org/github.com/gobs/args

## Example

    import "github.com/gobs/args"
    
    s := `one two three "double quotes" 'single quotes' arg\ with\ spaces "\"quotes\" in 'quotes'" '"quotes" in \'quotes'"`

	for i, arg := range GetArgs(s) {
		fmt.Println(i, arg)
	}

You can also "parse" the arguments and divide them in "options" and "parameters":

    import "github.com/gobs/args"

    s := "-l --number=42 -where=here -- -not-an-option- one two three"
    parsed := ParseArgs(s)

    fmt.Println("options:", parsed.Options)
    fmt.Println("arguments:", parsed.Arguments)

Or parse using flag.FlagSet:

    import "github.com/gobs/args"

    s := "-l --number=42 -where=here -- -not-an-option- one two three"
    flags := args.NewFlags("test")

    list := flags.Bool("l", false, "list something")
    num := flags.Int("number", 0, "a number option")
    where := flags.String("where", "", "a string option")

    flags.Usage()

    args.ParseFlags(flags, s)
	
    fmt.Println("list:", *list)
    fmt.Println("num:", *num)
    fmt.Println("where:", *where)
    fmt.Println("args:", flags.Args())

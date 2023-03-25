%{

package tasscript

import (
    "bytes"
	"log"
    "regexp"
    "strconv"
    "unicode"
	"unicode/utf8"
)

%}

%union {
    stmt Statement
    stmts Statements
    fp Floatpair
    cmd string
    name string
    dir direction
    loc Location
    lbegin loopBegin

    // A lot of the language uses ints, but because writing a lexer that has
    // enough context to know when to output them vs floats is hard, I modified the grammar
    // to use FLOATs everywhere and cast to int as needed
    f float32

    result  AST
}

%type <result> top
%type <stmt> statement start_cmd halt_cmd
%type <stmts> statements
%type <dir> direction
%type <loc> location
%type <lbegin> loop_begin

%type <fp> fpair

/* commands */
%token <token> BUILD CRAFT IDLE LAUNCH MINE PUT RECIPE ROTATE SPEED TAKE TECH WALK START LOCATION HALT LOOP ENDLOOP

/* directions */
%token <token> NORTH SOUTH EAST WEST

%token "\n"    NEWLINE
%token <f>     FLOAT
%token <name>  NAME

%%


top:
	start_cmd statements halt_cmd
    {
        a := AST{}
        if $1.Type != "" {
            a.Add(&$1)
        }
        for _, s := range $2 {
            if s.Type != "" {
                a.Add(s)
            }
        }
        if $3.Type != "" {
            a.Add(&$3)
        }
        tasParseResult = a.Clone()
        $$ = a
    }

start_cmd: /* empty */ {
}
|   START fpair NEWLINE
    {
        $$ = Statement{
            Type: "START",
            FloatVals: []float32{$2.X, $2.Y},
        }
    }

halt_cmd: /* empty */ {
    $$ = Statement{
            Type: "HALT",
    }
}
|   HALT NEWLINE
    {
        $$ = Statement{
            Type: "HALT",
        }
    }

statements: {

}
|   statement statements
    {
        st := append($$, $1.Clone())
        for _, s := range $2 {
            st = append(st, s.Clone())
        }
        $$ = st
    }

statement:
    NEWLINE {
        /* empty line */
    }
|   LOCATION NAME FLOAT FLOAT NEWLINE {
        $$ = Statement{
            Type: "LOCATION",
            StrVals: []string{$2},
            FloatVals: []float32{$3,$4},
        }
    }
|   BUILD location NAME direction NEWLINE {
        $$ = Statement{
            Type: "BUILD",
            Location: $2.Clone(),
            StrVals: []string{$3},
            Direction: $4,
        }
    }
|   MINE location FLOAT NEWLINE {
        $$ = Statement{
            Type: "MINE",
            Location: $2.Clone(),
            IntVals: []int{int($3)},
        }
    }
|   SPEED FLOAT NEWLINE {
        $$ = Statement{
            Type: "SPEED",
            FloatVals: []float32{$2},
        }
    }
|   RECIPE location NAME NEWLINE {
        $$ = Statement{
            Type: "RECIPE",
            Location: $2.Clone(),
            StrVals: []string{$3},
        }
    }
|   ROTATE location direction NEWLINE {
        $$ = Statement{
            Type: "ROTATE",
            Location: $2.Clone(),
            Direction: $3,
        }
    }
|   PUT location NAME FLOAT NAME NEWLINE {
        $$ = Statement{
            Type: "PUT",
            Location: $2.Clone(),
            StrVals: []string{$3, $5},
            IntVals: []int{int($4)},
        }
    }
|   TAKE location NAME FLOAT NAME NEWLINE {
        $$ = Statement{
            Type: "TAKE",
            Location: $2.Clone(),
            StrVals: []string{$3, $5},
            IntVals: []int{int($4)},
        }
    }
|   CRAFT NAME FLOAT NEWLINE {
        $$ = Statement{
            Type: "CRAFT",
            StrVals: []string{$2},
            IntVals: []int{int($3)},
        }
    }
|   TECH NAME NEWLINE {
        $$ = Statement{
            Type: "TECH",
            StrVals: []string{$2},
        }
    }
|   LAUNCH location NEWLINE {
        $$ = Statement{
            Type: "LAUNCH",
            Location: $2.Clone(),
        }
    }
|   IDLE FLOAT NEWLINE {
        $$ = Statement{
            Type: "IDLE",
            IntVals: []int{int($2)},
        }
    }
|   WALK location NEWLINE {
        $$ = Statement{
            Type: "WALK",
            Location: $2.Clone(),
        }
    }
|   LOOP loop_begin statements ENDLOOP NEWLINE {
        $$ = Statement{
            Type: "LOOP",
            Body: $3,
            IntVals: []int{$2.n},
        }
    }

loop_begin:
    FLOAT NEWLINE {
        $$ = loopBegin{
            n: int($1),
        }
    }

location:
    NAME {
        $$ = Location{
            Named: $1,
        }

    }
|   fpair {
        $$ = Location{
            Values: $1,
        }
    }

direction:
    NORTH {
        $$ = North
    }
|   SOUTH {
        $$ = South
    }
|   EAST {
        $$ = East
    }
|   WEST {
        $$ = West
    }


fpair:
    FLOAT FLOAT {
        $$ = Floatpair{X: $1, Y: $2}
    }
%%

// The parser expects the lexer to return 0 on EOF.  Give it a name
// for clarity.
const eof = 0

// The parser uses the type <prefix>Lex as a lexer. It must provide
// the methods Lex(*<prefix>SymType) int and Error(string).
type exprLex struct {
	input []byte
}

var identPat = regexp.MustCompile(`^[A-Za-z]([A-Za-z0-9-_]*[A-Za-z0-9])?$`) 

// The parser calls this method to get each new token
func (x *exprLex) Lex(yylval *tasSymType) int {
	add := func(b *bytes.Buffer, c rune) {
		if _, err := b.WriteRune(c); err != nil {
			log.Fatalf("WriteRune: %s", err)
		}
	}

    var (
        assign func(s string) bool
        ret int
    )
	
    for {
		c := x.next()
        var b = bytes.NewBuffer(nil)

        switch {
            case c == eof:
                return eof
            case c == '-' || (c >= '0' && c <= '9'):
                assign = func(s string) bool {
                    f, err := strconv.ParseFloat(s, 64)
                    if err == nil {
                        yylval.f = float32(f)
                        ret = FLOAT
                        return true
                    }
                    return false
                }
            case c >= 'A' && c <= 'Z':
                assign = func(s string) bool {
                    var ok bool
                    ret, ok = keywords[s]
                    yylval.cmd = s
                    return ok
                }
            case c >= 'a' && c <= 'z':
                assign = func(s string) bool {
                    if identPat.MatchString(s) {
                        yylval.name = s
                        return true
                    }
                    return false
                }
                ret = NAME
            case c == '\n':
                return NEWLINE
            case unicode.IsSpace(c):
                continue
            default:
                // syntax error
                log.Fatalf("invalid rune %q", c)
        }
        add(b, c)

        for {
            c = x.peek()
            if c == eof || unicode.IsSpace(c) {
                break
            }

            add(b, c)

            x.next()
        }

        s := b.String()

        if assign(s) {
            return ret
        }
        log.Printf("Invalid token %q for type %d", s, ret)
	}
}


// Return the next rune for the lexer.
func (x *exprLex) next() rune {
    return x.getNext(true)
}

func (x *exprLex) peek() rune {
    return x.getNext(false)
}

func (x *exprLex) getNext(advance bool) rune {
	if len(x.input) == 0 {
		return eof
	}
	c, size := utf8.DecodeRune(x.input)
    if advance {
        //log.Printf("Reading rune %q", c)
	    x.input = x.input[size:]
    }
	if c == utf8.RuneError && size == 1 {
		log.Print("invalid utf8")
		return x.next()
	}

	return c
}


// The parser calls this method on a parse error.
func (x *exprLex) Error(s string) {
	log.Fatalf("parse error: %s", s)
}


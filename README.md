# gocompiler
Хроленко Валерий Павлович, ДВФУ, 09.03.03ПИКД, 2022 год, Go
# Инструкции по запуску
Лексический анализатор
```
go run main.go -lex -source имя_файла
```
Парсер
```
go run main.go -ast -source имя_файла
```
# Реализуемое подмножество языка

### Обозначения
```
|   или
()  группировка
[]  опиональность (0 или 1 раз)
{}  повторение (0 или n раз)
+   Следование
```

```
Literal     = BasicLit | CompositeLit | FunctionLit 
BasicLit    = int_lit | float_lit | rune_lit | string_lit 
IdentifierList = identifier { "," identifier } 
CompositeLit  = LiteralType + LiteralValue 
LiteralType   = StructType | ArrayType  | SliceType | identifier [ TypeArgs ]

Expression = UnaryExpr | Expression + binary_op + Expression .
UnaryExpr  = PrimaryExpr | unary_op + UnaryExpr 

StructType    = "struct" + "{" + { FieldDecl ";" } + "}" 
FieldDecl     = (IdentifierList + Type | EmbeddedField) 
EmbeddedField = identifier [ TypeArgs ]

Type      = identifier + [ TypeArgs ] | TypeLit | "(" + Type + ")" 
TypeArgs  = "[" + TypeList + [ "," ] + "]" 
TypeList  = Type  { "," Type } 
TypeLit   = ArrayType | StructType | FunctionType | SliceType

ArrayType   = "[" + Expression + "]" + Type

SliceType = "["  +  "]" + Type 

FunctionType   = "func" +  Signature 

FunctionLit = "func" + Signature + Block 
Signature      = Parameters [ Result ]
Result         = Parameters | Type 
Parameters     = "(" + [ ParameterList + [ "," ] ] + ")" 
ParameterList  = ParameterDecl { "," ParameterDecl } 
ParameterDecl  = [ IdentifierList ] + Type 

Block = "{" + StatementList + "}" 
StatementList = { Statement ";" } 
Statement = Declaration | SimpleStmt | ReturnStmt | Block | IfStmt | ForStmt
SimpleStmt = Expression | IncDecStmt | Assignment | ShortVarDecl
IncDecStmt = Expression + ( "++" | "--" )
Assignment = ExpressionList + assign_op + ExpressionList
ShortVarDecl = IdentifierList + ":=" + ExpressionList 

ReturnStmt = "return" + [ ExpressionList ]

IfStmt = "if" + Expression + Block + [ "else" + ( IfStmt | Block ) ]

ForStmt = "for" + [ Expression | ForClause ] + Block
ForClause = [ SimpleStmt ] + ";" + [ Expression ] + ";" + [ SimpleStmt ]

Declaration   = ConstDecl | TypeDecl | VarDecl

VarDecl     = "var" + ( VarSpec | "(" + { VarSpec ";" } + ")" )
VarSpec     = IdentifierList + ( Type + [ "=" + ExpressionList ] | "=" + ExpressionList )


ConstDecl      = "const" ( ConstSpec | "(" + { ConstSpec ";" } + ")" ) 
ConstSpec      = IdentifierList [ [ Type ] + "=" + ExpressionList ] 
IdentifierList = identifier + { "," identifier } 
ExpressionList = Expression + { "," Expression } 

TypeDecl = "type" + ( TypeSpec | "(" + { TypeSpec ";" } + ")" ) 
TypeSpec = AliasDecl | TypeDef 
AliasDecl = identifier + "=" + Type
TypeDef = identifier + [ TypeParameters ] + Type
TypeParameters  = "[" + TypeParamList [ "," ] + "]" 
TypeParamList   = TypeParamDecl + { "," + TypeParamDecl } 
TypeParamDecl   = IdentifierList + TypeElem
TypeElem       = Type { "|" + Type } 

VarDecl     = "var" + ( VarSpec | "(" + { VarSpec ";" } + ")" ) 
VarSpec     = IdentifierList + ( Type + [ "=" + ExpressionList ] | "=" + ExpressionList )

CompositeLit  = LiteralType + LiteralValue
LiteralType   = StructType | ArrayType | identifier
LiteralValue  = "{" + [ ElementList [ "," ] ] + "}" 
ElementList   = KeyedElement { "," KeyedElement } 
KeyedElement  = [ Key + ":" ] + Element 
Key	      = identifier | Expression | LiteralValue
Element	      = Expression | LiteralValue

PrimaryExpr = Operand | MethodExpr 
| PrimaryExpr + Selector | PrimaryExpr + Index | PrimaryExpr + Arguments
Selector       = "." + identifier 
Index          = "[" + Expression [ "," ] + "]"
Arguments      = "(" + [ ( ExpressionList | Type + [ "," + ExpressionList ] ) ] + ")"
MethodExpr    = Type + "." + identifier
Operand     = Literal | identifier + [ TypeArgs ] | "(" + Expression + ")"
```

" Vim syntax file
" filetype: ledger
" by Chris Howey

syn match ledgerComment /;.*/

syn match ledgerFloatAmount /\<-\?\d\+\.\d\+\>/
syn match ledgerIntAmount /\<-\?\d\+\>/

syn match ledgerAccount /\<\(\(\w\|\s\)\+:\)\+\(\w\|\s\)\+\D\>/
syn match ledgerAccountEOL /\<\(\(\w\|\s\)\+:\)\+\w\+\>$/

syn match ledgerPayeeComment /;.*/ contained
syn match ledgerDate /^\d\{4}\(\/\|-\)\d\{2}\(\/\|-\)\d\{2}\>/ contained

syn match ledgerTopline /^\S.*/ contains=ledgerDate,ledgerPayeeComment

syn region ledgerFold start="^\S" end="^$" transparent fold

highlight default link ledgerDate Function

highlight default link ledgerComment Comment
highlight default link ledgerPayeeComment Comment
highlight default link ledgerFloatAmount Number
highlight default link ledgerIntAmount Number

highlight default link ledgerAccount Identifier
highlight default link ledgerAccountEOL Identifier

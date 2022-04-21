" Vim syntax file
" filetype: ledger
" by Chris Howey

syn match ledgerComment /;.*/

syn match ledgerFloatAmount /\<-\?\d\+\.\d\+\>/ contained
syn match ledgerIntAmount /\<-\?\d\+\>/ contained

syn match ledgerAccount /\<\%(\%(\w\|\s\)\+:\)\+\%(\w\|\s\)*\w\+\s\{2}/me=e-2 contained
syn match ledgerAccountEOL /\<\%(\%(\w\|\s\)\+:\)\+\%(\w\|\s\)*\w\+\>$/ contained

syn match ledgerPostingEmpty /^\s\{2,}.*$/ contains=ledgerAccountEOL
syn match ledgerPostingAmount /^\s\{2,}\S.\+\s\{2,}.\+$/ contains=ledgerAccount,ledgerFloatAmount,ledgerIntAmount

syn match ledgerPayeeComment /;.*/ contained
syn match ledgerDate /^\d\{4}\%(\/\|-\)\d\{2}\%(\/\|-\)\d\{2}\>/ contained

syn match ledgerTopline /^\S.*$/ contains=ledgerDate,ledgerPayeeComment

syn region ledgerFold start="^\S" end="^$" transparent fold

highlight default link ledgerDate Function

highlight default link ledgerComment Comment
highlight default link ledgerPayeeComment Comment
highlight default link ledgerFloatAmount Number
highlight default link ledgerIntAmount Number

highlight default link ledgerAccount Identifier
highlight default link ledgerAccountEOL Identifier

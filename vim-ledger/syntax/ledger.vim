" Vim syntax file
" filetype: ledger
" by Chris Howey

syn match ledgerComment /;.*/

syn match ledgerAmount /\s\{2}-\?\d\+\.\d\+$/ms=s+2
syn match ledgerAccount /^\s\{4}.*\s\{2}/ms=s+4,me=e-2

syn match ledgerDate /^\d\{4}\%(\/\|-\)\d\{2}\%(\/\|-\)\d\{2}\s/me=e-1

syn region ledgerFold start="^\S" end="^$" transparent fold

highlight default link ledgerDate Function
highlight default link ledgerComment Comment
highlight default link ledgerAmount Number
highlight default link ledgerAccount Identifier

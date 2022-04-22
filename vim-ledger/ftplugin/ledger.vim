" Vim filetype plugin file
" filetype: ledger
" by Chris Howey

setl omnifunc=LedgerComplete

if !exists('g:ledger_main')
  let g:ledger_main = '%:p'
endif

if !exists('g:ledger_bin') || empty(g:ledger_bin)
	if executable('ledger')
		let g:ledger_bin = 'ledger'
	endif
elseif !executable(g:ledger_bin)
	unlet! g:ledger_bin
	echohl WarningMsg
	echomsg 'Command set in g:ledger_bin is not executable'
	echohl None
endif

if !exists('g:ledger_accounts_cmd')
  if exists('g:ledger_bin')
    let g:ledger_accounts_cmd = g:ledger_bin . ' -f ' . shellescape(expand(g:ledger_main)) . ' accounts'
  endif
endif

function! LedgerComplete(findstart, base)
	if a:findstart
	    let line = getline('.')
		let end = col('.') - 1
		let start = 0
		while start < end && line[start] =~ '\s'
	      let start += 1
	    endwhile
	    return start
	else
	    let res = []
	    for m in systemlist(g:ledger_accounts_cmd . ' -m "' . a:base . '"')
			call add(res, m)
	    endfor
	    return res
	endif
endfun

function! _LedgerFormatFile()
	if exists('g:ledger_bin') && exists('g:ledger_autofmt_bufwritepre') && g:ledger_autofmt_bufwritepre
		let substitution = system(g:ledger_bin . ' print -f -', join(getline(1, line('$')), "\n"))
		if v:shell_error != 0
			echoerr "While formatting the buffer via fmt, the following error occurred:"
			echoerr printf("ERROR(%d): %s", v:shell_error, substitution)
		else
			let [_, lnum, colnum, _] = getpos('.')
			%delete
			call setline(1, split(substitution, "\n"))
			call cursor(lnum, colnum)
		endif
	endif
endfunction

if has('autocmd')
	augroup ledger_fmt
		autocmd BufWritePre * call _LedgerFormatFile()
	augroup END
endif

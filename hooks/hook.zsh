_synesthesia_hook() {
  synesthesia
}
typeset -ag chpwd_functions;
if [[ -z ${chpwd_functions[(r)_synesthesia_hook]} ]]; then
  chpwd_functions=( _synesthesia_hook ${chpwd_functions[@]} )
fi

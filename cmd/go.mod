module github.com/kkonat/WoL/cmd

require (
  github.com/kkonat/WoL/Internal/wol v0.0.0
)

replace github.com/kkonat/WoL/Internal/wol => ../Internal/wol
go 1.16

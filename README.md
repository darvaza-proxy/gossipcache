Gossip powered group cache library
==================================

Why not plain groupcache?
-------------------------
Manually managing the list of peers sucks

Why mailgun's groupcache?
-------------------------
We using mailgun's groupcache fork instead of Google's 
because they've added support for `context.Context`, go.mod,
and explicit key removal and expiration.

See also
--------
* https://github.com/hashicorp/memberlist
* https://github.com/mailgun/groupcache
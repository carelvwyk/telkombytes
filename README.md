Telkom Bytes
=

Telkom does not provide a nice interface to track Telkom Mobile internet usage over time. Current usage information (bytes remaining on data cap) is available from an unauthenticated web portal when visited from a Telkom Mobile internet connection or from an authenticated web portal.

I initially started off by trying to access the unauthenticated interface but ran into some problems with infinite redirects, so I turned to the authenticated portal.  I noticed that I could simulate a login and retrieve bundle balances using curl but not with Go. After some time trying to figure out what is going on, I realised that Telkom provides cookies with invalid names (contains a colon). Go is a lot more strict than curl and web browsers when it comes to cookie names, so it was rejecting the important Telkom context cookies.

If you look at the code you'll see I replace the colon with the word "COLON" when the cookie gets added to the cookiejar and then the reverse transform is applied when requests are created with cookies from the cookiejar. 

My Telkom mobile service balances are pushed to AWS CloudWatch metrics for graphing and alerting. 

This project is Go based because I love Go.
#!/bin/sh

# Start fail2ban
[ $XUI_ENABLE_FAIL2BAN == "true" ] && fail2ban-client -x start

# Run 4y-ui
exec /app/4y-ui

can I enable application-wide logging for all my queries?
It is not possible. Logging must be explicitly enabled for each and every query in order for the query to be logged. This is because there is no global database object provided by the library that allows for the configuration of a global logger.

can I have levelled logging?
sq's logger is intentionally dumb and straigtforward, it does not support levelled logging. You will have to implement your own.

# bbdb

Bbdb is a little utility for reading drive stats from [BackBlaze](https://www.backblaze.com/), and
putting them into a sqlite3 database.  BackBlaze is unique among cloud backup/storage providers, in that they are transparent about their operational statistics.  They make SMART data available for all the drives they use, and by downloading that data into a database, amateurs like me can analyze the data for useful things like the reliability of various brands and models of drives.

There's also a little shell script for downloading the stats.  This may stop working if BackBlaze changes the format of their [download page](https://www.backblaze.com/b2/hard-drive-test-data.html).

To use bbdb, first run `download.sh` in an empty directory, then extract the contents of all the zip files.  This will give you a bunch of directories with a bunch of `.csv` filies containing all the data.  Then run `bbdb <data_dir>`.  This will recursively search for data files under `data_dir`, and add all the data to `drive_stats.db`, which is a SQLite3 database.  It also creates a SQL view (based on BackBlaze's own sample code) to give you a reliability stat for all the drive models in the database.

After that, you're pretty much on your own to come up with insights from the data. :)

The download script and bbdb itself are idempotent.  That is, you can run them again with the same inputs, and they will incorporate new data without erasing the old.  `download.sh` will only download files that aren't already downloaded, and bbdb will only insert new data into the database.  (Be warned, thoug, bbdb will spit out errors for the rows it's skipping.)

Although bbdb incorporates a few snippets of SQL code from BackBlaze's own sample scripts, this project is in no way affiliated with the company.  Don't blame them if there's a problem with this code.

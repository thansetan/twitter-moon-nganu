# How it works
1. User authenticate the app,
2. The app will create a cronjob on [cron-job.org](https://cron-job.org),
3. The cronjob will run every hour at the 31st minute,
4. The cronjob will send a request to [this app](https://github.com/thansetan/twitter-moon) at the /picture endpoint and it will change the user's profile picture to the current moon phase picture retrieved from [this page](https://svs.gsfc.nasa.gov/5048).

# Why not just use time.Ticker and goroutines?
These things can actually be done using time.Ticker and goroutines, but I don't want to run this app on my own server, so I use [cron-job.org](https://cron-job.org) to do the job for me 👍🏻.

# Why not write the code for changing the profile picture in Go?
I'm too lazy to do that, so I just use [this app](https://github.com/thansetan/twitter-moon) that I wrote in Python.
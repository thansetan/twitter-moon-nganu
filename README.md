# How it works
1. User authenticate the app,
2. The app will create a cronjob on [cron-job.org](https://cron-job.org),
3. The cronjob will run every hour at the 32nd minute,
4. The cronjob will send a request to [this app](https://github.com/thansetan/twitter-moon) at the /picture endpoint and it will change the user's profile picture to the current moon phase picture retrieved from [this page](https://svs.gsfc.nasa.gov/5048).
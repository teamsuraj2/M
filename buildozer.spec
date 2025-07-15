[app]

# (str) Title of your application
title = HelloVivek

# (str) Package name
package.name = hellovivek

# (str) Package domain
package.domain = org.example.vivek

# (list) Only include Python files (since you don't use images or KV files)
source.include_exts = py

# (str) Your main file
main.py = main.py
source.dir = .

# (str) App version
version = 1.0

# (list) Dependencies required
requirements = python3,kivy

# (str) Supported orientation
orientation = portrait

# (bool) Run app in fullscreen
fullscreen = 1

# (str) Android SDK/NDK settings
android.api = 31
android.minapi = 21
android.ndk = 25b

# (str) Architecture
android.archs = armeabi-v7a

# (bool) Use private storage
android.private_storage = True

# (str) Optional: set your app icon (only if you have one)
# icon.filename = icon.png
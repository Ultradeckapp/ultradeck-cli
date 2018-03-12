

# TODO next

* [x] Implement pull
* [x] Implement updated_at timestamp checking for push + pull
* [x] create temporary security creds on app.ultradeck.co for CLI file uploads
* [x] see TODOs in asset_manager to finish it up
* [x] asset push / sync
* [x] pull assets should pull remote assets only, not push local ones
* [x] push assets should push local assets only, not pull remote ones
* [ ] Force command to force-push or force-pull
* [ ] Implement watch
* [ ] webhook support on frontend to support auto-updating
* [ ] Fix frontend (if needed b/c of potential breaking api changes)

# TODO once almost done

* [ ] homebrew
* [ ] other (legit) sites that host binaries

---

If pushing:

* updated_at Timestamp on client must be equal to or greater than updated_at timestamp on server
* will need to do a GET request check for that

If pulling

* updated_at on SERVER must be equal to or greater than updated_at timestamp on server
* updated_at on

---

# Aug 24 2017

* push now works.
* should now be easy to implement pull and watch, now that we're centering around pushing + pulling .ud.json.
* [ ] the frontend is probably broken, because I changed the api sig on the backend to support the cli.
* upgrade account works well (maybe rename "upgrade" to something else.)

# Aug 29 2017

* Aws temporary S3 creds working!
* got a file to successfully upload with temp creds.

# Aug 30

* Got file syncing algorithm mostly hammered out, although it is not tested very well
  * test uploading file on web view and see if it downloads locally
  * web view is all fucked up, needs a lot of love
* s3 is not uploading with public read, although I fixed it but did not test it
* SyncAssets needs to be part of push (or pull?) need to think about this more.  right now it is an external command.

# Aug 31

* Separated out pulling remote assets and pushing local assets.  I don't see a situation where these should be run together with a single command.

# Sep 1

how do I definitively delete an asset?

if pushing and the remote asset is there, but the local asset is not there, then prompt the user
if pulling and the remote asset is not there, but the local asset is there, then do nothing.

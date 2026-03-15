# Zero to Launch

This is a set of high level steps with some notes on gotchas and lore. Between this and AGENTS.md you can walk through getting set up from complete scratch with the help of any free AI chat-based tool. If you're running an agentic coding assistant just have it read this file and AGENTS.md and ask it to guide you through the process.
Either way, that workflow will provide up to date details specific for your needs much better than a document I wrote that "worked for me" however long ago.

This branch of the repo contains the source for a Cloudflare Worker project and the tooling for a Markdown -> Web 1.0 HTML workflow. You'll need to make edits and do some renaming for it to work for the domain name(s) you choose. Domain purchase comes after confirming the setup works — so you're not spending money before knowing everything runs.

Here's the steps. Follow them in order. At the other end is your new site.

## Install tools
The tools listed below are all that are required, but it is highly recommended that you use an AI agent to help guide you through the process and to help with the nitty gritty details.

### Node.js / npm (Node Package Manager)
* Needed to install and run the Cloudflare Worker tool ```wrangler```
* It's available on all major platform package managers: macOS (brew), Linux (apt or pacman), and Windows (winget)
* Install or upgrade to the latest LTS version - wrangler requires version 18+ (as of this writing).

### Wrangler
* Command line tool for developing and managing Cloudflare Workers
* Install it globally with ```npm install -g wrangler```

### A Code Editor
* I really like [Zed](https://zed.dev). Wicked fast, lots of great stuff out of the box, free, lots of good extensions available. See instruction installations on their site.
* VS Code is also very popular

### Command Line Git Tools and Git LFS
Your code editor likely has built-in git support, but the command line tools will provide options that may help you get out of corners. There's a lot of recipes for using them both online and trained into AI, much more so than for GUI tools.
You need git LFS (Large File Storage) to handle pulling the pre-built site-tool binaries (with git LFS installed it will be automatic)
* macOS
  * ```xcode-select --install```
    * try ```git lfs install``` - if this succeeds or says it's already installed, you're done
  * (if not included via xcode-select) ```brew install git-lfs && git lfs install```
* Linux: ```apt install git git-lfs``` or ```pacman -S git-lfs``` then ```git lfs install```
* Windows: ```winget install -e --id Git.Git``` (git LFS is included)
  * IMPORTANT: Any terminal commands specified in this document are to be run from a git bash terminal

### Gnu Make
* macOS: you're set from installing the xcode command line tools
* Linux: ```make``` is available on package managers (apt or pacman) and is a dependency of many dev packages (e.g. "build-essential")
* Windows: ```winget install -e --id GnuWin32.Make```

### Pandoc
This is the tool that converts Markdown to HTML for the .md -> .html web 1.0 authoring workflow
* macOS: ```brew install pandoc```
* Linux: ```sudo apt install pandoc``` or ```sudo pacman -S pandoc```
* Windows: ```winget install -e --id JohnMacFarlane.Pandoc```

### ImageMagick
This is used in the converting .svg images to .ico format for the Safari fallback version of the site favicon.
* macOS: ```brew install imagemagick```
* Linux: ```sudo apt install imagemagick``` or ```sudo pacman -S imagemagick```
* Windows: ```winget install -e --id ImageMagick.ImageMagick```
IMPORTANT: on macOS, the first step of the process (.svg -> .png) is done by qlmanage, which has a better SVG renderer than ImageMagick. On Windows and Linux ImageMagick does the entire conversion. You'll want to check your results when updating the favicon. (A message will print to the console reminding you to do so.)


## Create GitHub account (skip if you already have one appropriate for this project)

## [Create a new repo from this template](https://github.com/new?template_name=personal-site-for-cost-of-domain&template_owner=benn-herrera) (link will start the process)
  * Check "Include all branches" (or you'll just get the skeleton project on "main")
  * Choose a repo name indicating it is a worker project (e.g. personal-site-worker) 
    * Don't name it after your domain - this could easily get confusing or out of date
    * You can actually back multiple domains from one worker project
  * Clone it locally - for doc purposes we'll assume you cloned it to ```~/projects/personal-site-worker``` (or the Windows equivalent)
  * Overwrite the state of branch "main" with branch "markdown-web-1.0-workflow"
    * ```git checkout main && git checkout -b main-backup && git checkout main``` - make a backup of the bare bones main in case you need it, switch back to main
    * ```git reset --hard markdown-web-1.0-workflow``` - force the state of "main" to that of "markdown-web-1.0-workflow"
    * ```git push -f origin main``` - force push the updated "main" branch to GitHub
    * ```git branch -D markdown-web-1.0-workflow``` - delete the tooled up workflow source branch - you're using it now on main

## Initial Setup and Workflow
  
### Confirm it works out of the box
* Open a terminal and change directory to your project clone
* Run ```make -C maint serve``` in a terminal - this runs `wrangler dev`
  * The startup banner in the terminal will show the local URL your site is available at (e.g. `http://localhost:8787` or `http://192.168.4.55:8787`)
* Hit 'b' in the terminal window to pop your browser to the local site
  * Your browser should be open showing the placeholder content for "My Personal Site"
* From a terminal run ```cd public/my-personal-site.me && make -B```
  * The `-B` flag forces a full rebuild regardless of file timestamps — used here to confirm the pipeline runs cleanly. Normal day-to-day use is just `make`.
  * Check that it completes with no errors. Ideally no files change in git — meaning the pipeline reproduced the existing output exactly.
    * Caveat: if pandoc has changed version the output might differ slightly but render exactly the same.

### See how multi-site support works (optional - skip if you're only going to serve one domain)
If you want to serve multiple sites (domains) from this worker you'll need to know how this works.
AGENTS.md has details, but here's how to test it out of the box.
* In your browser after the URL in the address bar `http://localhost:8787` add a query parameter so it looks like this: `http://localhost:8787?d=my-second-personal-site.me` and hit 'enter'
* You should see the title and text change to "My Other Personal Site".

The `d` (for domain) query parameter is "sticky" (via a cookie) - after you use it once that's the site it will stay on until you change it back or launch a new browser.

You'll notice if you edit the URL and remove the `?d=my-second-personal-site.me` and hit 'enter' it will stay on "My Other Personal Site" - that's the sticky behavior.

You can switch back by changing the URL to `http://localhost:8787?d=my-personal-site.me`, and if you remove the query, again, it will stay with the last site you specified.

### Buy your domain from Cloudflare (skip if you already have one)
IMPORTANT: Do not buy a domain from Wix, Squarespace or any of the one-stop-shops. They may lock you into their service and make it impossible to use Cloudflare for 60 days (plus the hassle of moving the domain). They also charge a markup for domain registration, which Cloudflare does not.
  * Depending on what top level domain you choose (.me, .com, .dev, .ai, etc.) the price can vary significantly.
  * The standard choice for a personal site is .me
  * You may also want to add email handling
    * You won't get an inbox, but you can forward messages sent to name@your-domain.me to get forwarded to an existing email address.
    * You can add separate addresses by name or create a catch-all address that forwards all email to a single inbox.
  * Ask your AI assistant about the security options to change - the defaults are frequently too permissive

### Use a domain you already own (skip if you bought it from Cloudflare)
The easiest path will be to transfer the domain to Cloudflare. If you bought a domain from Wix or similar service that won't let you change DNS "NS" (name server) records, you're stuck until 60 days from time of registration, when the transfer lockout period lapses. If you used some other registrar it may be possible to modify their DNS settings to work with Cloudflare, but there's no guarantee.
  * Ask your AI assistant to help with this
  * Cloudflare's site also provides some guidance
  * This is definitely the trickiest path
    * In all seriousness, best of luck
    * Plumbing the nest of reasons things might have gone wrong is zero fun.

### Change the names to match your project
IMPORTANT: If you skipped the domain purchase step and are not sure what your domain name will be you *can* continue, but the step of plugging in your real domain name(s) will need to be done. It's easier to get it done and out of the way here so you can get everything fully planned before initial publication.

* Edit `wrangler.jsonc` and change the "name" field to match your repo name
  * You'll use this name later when setting up the Worker on Cloudflare
  * It's not required, but it keeps things clear when the Worker and repo names match
* Under `public/` rename `my-personal-site.me` - change `my-personal-site.me` to the domain name you want to serve
  * If you don't have a domain name picked out, don't worry this is easy to change later
  * If you're only going to serve one domain, delete `public/my-second-personal-site.me`
* If you're going to serve multiple domains
  * rename `my-second-personal-site.me` to your 2nd domain
  * copy `my-personal-site.me` under `public/` to a new directory for each domain you want to serve
  * The `sites` list in `src/index.js` is maintained automatically — running `make` from any site directory updates it to match the current directories in `public/`. No manual edits needed.
* Edit the placeholder `index.md` files in your site content roots to reflect the new names
* Verify everything is working as expected
  * for each domain directory under public you've set up run ```make -C public/<domain-name.tld>```
  * Hit 'q' in the `wrangler dev` terminal to stop the server
  * Re-run ```make -C maint serve``` and refresh your browser
    * NOTE: this is only because wrangler.jsonc was changed - not part of normal workflow
    * You may need to do a hard refresh (Ctrl+Shift+R on Windows/Linux, Cmd+Shift+R on macOS) if you're still seeing the old content

**Now ask your coding agent to finish the rename.** The build pipeline has placeholder names spread across several more files that the agent can handle efficiently. Give it the actual values for your setup and ask it to make the following changes:

* `maint/common.mk` — update `SITE_AUTHOR` to your name
* The domain name is inferred automatically from the directory name — no variable to update in any Makefile.
* `public/**/*.template.html` — templates co-located with their pages contain hardcoded placeholder domain names (`my-personal-site.me`, `my-second-personal-site.me`) in `<title>` tags, RSS link titles, nav name-marks, and cross-site nav links. The agent should grep for these and replace with your actual domains.
* `public/**/*.md` — the content markdown files contain placeholder domain names in cross-site link text. The agent should grep for `my-personal-site.me` and `my-second-personal-site.me` and update any links or references to match your actual domains.
* `AGENTS.md` — update the domain and make target name references throughout to match your actual setup.

After the agent is done, run `cd public/my-personal-site.me && make` (and `cd public/my-second-personal-site.me && make` if using a second site) to rebuild everything and verify there are no errors.

* Commit your changes to "main" with git and push them to GitHub

## Work on your site until you have something you are willing to publish
You (and your AI assistant) can go nuts here or you can just pick a color scheme and make a "Coming Soon" page.
General advice:
* See AUTHORING_README.md for the guide on using the Markdown -> Web 1.0 tooling
* Make the aesthetic choices your own - don't let AI pick it all for you
* The nav and general structure are a starting point - tear it up and make something you like
* Fold, spindle, and mutilate the structure, tooling, and whatever else you need to in order to make it all work best for you
* Don't forget to commit your changes to "main" and push them to GitHub

## Create Cloudflare account and sign in
* You can, but don't have to choose the "Use GitHub Account" option when creating your Cloudflare account to save some steps.

## Create a Worker project and attach your repo
* Create a new Worker project - make sure it is a Worker and *not* Pages
  * IMPORTANT: `workers.dev` subdomain selection
    * during this process you'll be asked to choose a `workers.dev` subdomain name
    * this name will be used for all Worker project deployment URLs going forward, e.g. `personal-site-worker.[my-subdomain].workers.dev`
    * this can be changed later but it will break existing routing and you'll have to do some cleanup
    * recommendation: use your Cloudflare user name as the subdomain
    * `personal-site-worker.myusername.workers.dev` will be clear and unambiguous
  * Choose the GitHub option when creating your Worker project
  * You'll need to allow Cloudflare access to the repo you're attaching
  * Their dialog will walk you through the steps - unless you have a specific reason otherwise, limit the access to just your personal-site-worker repo when the option comes up for one repo or all your repos
* IMPORTANT: By default your Worker will automatically stay in sync with the "main" branch of your repo
  * See [Deployment](#deployment) for more information
  * Changes on the "main" branch pushed to GitHub will be visible to the world
  * See [Post Setup Workflow](#post-setup-workflow) for details on creating safe changes and vetting them before deploying
* At this point `personal-site-worker.my-cloudflare-user-name.workers.dev` will be visible to the world
  * `workers.dev` is the domain for deployed Cloudflare Workers - your domain name will alias to that address
  * See your settings for the Worker for details. At the time of this writing that information is at the top of the Settings page.

## Catch your breath a second
You'll note that up to now you've had to create probably one, maybe two accounts, both on free tiers of service.
You have a public corner on the internet. It has a name you don't fully control, but does have your name attached to it via your Cloudflare account name.

The last steps here are for connecting that public presence to the domain name(s) you chose.

## Connect your worker to a domain

### Add the domain to the list of domains your worker backs
  * On your Worker's Settings page there's a button for adding "custom domains"
  * Add your bare domain name (e.g. `my-personal-site.me`)
  * Add `www.my-personal-site.me`
    * Cloudflare should automatically create the CNAME DNS record for you
    * If you are using a separate registrar, ensure you have a CNAME alias that points `www.your-domain.me` to `your-domain.me`
  * Verify that you're live by browsing to `https://my-personal-site.me` and `https://www.my-personal-site.me`
    * It might take a minute for the DNS records to update and the site to become visible

## Congratulations, you're live!
That was the hardest part. Now you can focus on aesthetic decisions, content, writing, and honing.
Now, all that's left is everything - but it's your everything.

## Post-Setup Workflow 
* remove the old README.md and rename AUTHORING_README.md to README.md - this will be your work reminders.
* After you're done with the initial setup and perusing this guide, you can delete it as well. All the information you'll need going forward is in AGENTS.md and your new README.md
* If you're using a coding agent have it update AGENTS.md for you with the changes you've made. Keeping that up to date will save you a lot of trouble (and tokens).

You're going to want a reasonable workflow that lets you make changes and try them out without the world seeing every typo and change of direction. The primary workflow is local - ```wrangler dev```, make your edits, check on your machine and other devices before deploying. If you want to get feedback from a select group of people or test things that have elements that need to be checked in a real environment you'll want to use preview deployments.

If you are not familiar with using git branches and merging back to main, I highly recommend you look into it. Any AI chatbot can give you the hows and whys.

### Deployment
By default deployment is handled automatically. Any time you push changes to your main branch on GitHub your Cloudflare Worker will re-deploy using the new state. This usually only takes 30 seconds or so. Occasionally the auto-deploy mechanism will get stuck. You can disable this behavior and limit deployments to manually triggered from the web interface or using wrangler.

Manual deployment via wrangler:
* ```wrangler deploy```
  * NOTE: you'll need to do ```wrangler login``` first - it will open your browser and may ask you to do some steps.
  * After that you'll be logged in semi-permanently
* wrangler has no knowledge of your git repo - it will use what's present exactly as it is
  * NOTE: this will use the exact state of the local files in your working directory at the time you invoke the command
  * Make sure everything is committed (in your history) and ready for public visibility
    * This is not for wrangler, it is so that you don't ever have a deployed state you can't recover via git 

### Preview Deployment
Cloudflare provides a way of previewing changes in a deployed environment before deploying for your live site (production).
This is done by just creating a branch and pushing it to GitHub. 
* The URL for previewing that branch will be `https://[working-branch-name]-personal-site-worker.my-cloudflare-user-name.workers.dev`
* ^^^ may change - check your Cloudflare Worker settings - it specifies the URLs for production and preview deployments

By default this feature is enabled, but it can be disabled in Worker settings. You can also make these preview deployments visible only to you and authorized users - ask your AI assistant about it.

// @ts-check
// @ts-ignore
const initMediaFiles = [{{ StringsJoin .MediaFiles ", " }}];
const transitionVideoPath = {{ JSEscape .TransitionVideo }};
const playOnlyOne = {{ .PlayOnlyOne }};
const loopFirstVideo = {{ .LoopFirstVideo }};
const hashKey = {{ JSEscape .HashKey }};
// @ts-ignore
const isOBS = !!(window?.obsstudio?.pluginVersion);

/**
 * VideoPlayerManager handles the dual-player system for seamless video playback
 */
class VideoPlayerManager {
  /**
   * @param {string} player1Id - ID of the first video element
   * @param {string} player2Id - ID of the second video element
   */
  constructor(player1Id, player2Id) {
    /** @type {HTMLVideoElement} */
    this.player1 = /** @type {HTMLVideoElement} */ (document.getElementById(player1Id));
    /** @type {HTMLVideoElement} */
    this.player2 = /** @type {HTMLVideoElement} */ (document.getElementById(player2Id));
    /** @type {boolean} */
    this.isTransition = true; // set true for init on purpose
    /** @type {PlaylistManager} */
    this.playlistManager = new PlaylistManager(initMediaFiles, hashKey);
    /** @type {number|null} */
    this.preloadTimer = null;
    /** @type {Function[]} */
    this.eventListenerCleanup = [];

    this.setupEventListeners();
  }

  /**
   * Clean up all event listeners and timers
   */
  cleanup() {
    // Clear any pending timer
    if (this.preloadTimer) {
      clearTimeout(this.preloadTimer);
      this.preloadTimer = null;
    }

    // Clean up all event listeners
    this.eventListenerCleanup.forEach(cleanupFn => cleanupFn());
    this.eventListenerCleanup = [];
  }

  /**
   * Set up event listeners for both players
   */
  setupEventListeners() {
    // Player 1 ended event
    const player1EndedHandler = () => {
      if (!loopFirstVideo) {
        this.playlistManager.progressPlaylist();
      }
      this.playNext(this.player2, this.player1);
      // Preload the next video after this one starts playing
      if (this.preloadTimer) clearTimeout(this.preloadTimer);
      this.preloadTimer = setTimeout(() => this.preloadNextVideo(), 1000);
    };
    
    this.player1.addEventListener('ended', player1EndedHandler, { passive: true });
    this.eventListenerCleanup.push(() => {
      this.player1.removeEventListener('ended', player1EndedHandler);
    });

    // Player 2 ended event
    const player2EndedHandler = () => {
      if (!loopFirstVideo) {
        this.playlistManager.progressPlaylist();
      }
      this.playNext(this.player1, this.player2);
      // Preload the next video after this one starts playing
      if (this.preloadTimer) clearTimeout(this.preloadTimer);
      this.preloadTimer = setTimeout(() => this.preloadNextVideo(), 1000);
    };
    
    this.player2.addEventListener('ended', player2EndedHandler, { passive: true });
    this.eventListenerCleanup.push(() => {
      this.player2.removeEventListener('ended', player2EndedHandler);
    });

    // Add error handling for both players
    this.setupErrorHandling(this.player1);
    this.setupErrorHandling(this.player2);

    // Add controls for non-OBS environments
    if (!isOBS) {
      this.player1.setAttribute('controls', 'true');
      this.player2.setAttribute('controls', 'true');
    }
  }

  /**
   * Set up error handling for a player
   * @param {HTMLVideoElement} player - The video player element
   */
  setupErrorHandling(player) {
    const sources = player.querySelectorAll('source');
    sources.forEach(source => {
      const errorHandler = this.handleMediaError.bind(this);
      const stalledHandler = this.handleMediaError.bind(this);
      const suspendHandler = this.handleMediaError.bind(this);
      const waitingHandler = this.handleMediaError.bind(this);

      source.addEventListener('error', errorHandler);
      source.addEventListener('stalled', stalledHandler);
      source.addEventListener('suspend', suspendHandler);
      source.addEventListener('waiting', waitingHandler);

      // Store cleanup functions
      this.eventListenerCleanup.push(() => {
        source.removeEventListener('error', errorHandler);
        source.removeEventListener('stalled', stalledHandler);
        source.removeEventListener('suspend', suspendHandler);
        source.removeEventListener('waiting', waitingHandler);
      });
    });
  }

  /**
   * Handle media errors
   * @param {ErrorEvent | Event} event - The error event
   */
  handleMediaError(event) {
    const { type, target } = event;
    const videoSrc =
      // @ts-ignore
      /** @type {string} */ (target?.src) || 'UNKNOWN SOURCE FILE ERROR';

    switch (type) {
      case 'error':
        const errorStr = `Error loading: ${videoSrc}`;
        console.error(errorStr);
        if (isOBS) break;

        // Avoid DOM thrashing by only replacing if error element doesn't exist
        let errorElement = document.getElementById('error');
        if (!errorElement) {
          const errorText = document.createTextNode(errorStr);
          errorElement = document.createElement('h1');
          errorElement.setAttribute('id', 'error');
          errorElement.appendChild(errorText);

          // Hide video elements instead of clearing entire body
          const videoElements = document.querySelectorAll('video');
          videoElements.forEach(video => video.style.display = 'none');

          document.body.appendChild(errorElement);
        } else {
          // Update existing error message
          errorElement.textContent = errorStr;
        }
        break;
      case 'stalled':
      case 'suspended':
      case 'waiting':
      default:
        console.error(`Error loading: [TYPE: ${type}] ${videoSrc}`);
    }
  }

  /**
   * Preload the next video in the playlist
   */
  preloadNextVideo() {
    if (playOnlyOne) return; // Don't preload if we're only playing one video

    // Get the next video but don't remove it from the playlist yet
    const nextVideoSrc = this.playlistManager.getNextItem(true);
    if (!nextVideoSrc) return;

    // Create a preload link for the browser to fetch the video in advance
    const preloadLink = document.createElement('link');
    preloadLink.rel = 'prefetch'; // Use prefetch instead of preload for video
    preloadLink.href = nextVideoSrc;

    // Remove any existing prefetch links to avoid duplicates
    const existingPrefetch = document.querySelectorAll('link[rel="prefetch"]');
    existingPrefetch.forEach(link => link.remove());

    // Add the new preload link
    document.head.appendChild(preloadLink);
    console.debug('Preloading next video:', nextVideoSrc);
  }

  /**
   * Play the next video
   * @param {HTMLVideoElement} currentPlayer - The currently playing video element
   * @param {HTMLVideoElement} nextPlayer - The next video element to play
   */
  playNext(currentPlayer, nextPlayer) {
    try {
      const currentMp4Source = /** @type {HTMLSourceElement} */(currentPlayer.querySelector('source'));
      const nextMp4Source = /** @type {HTMLSourceElement} */(nextPlayer.querySelector('source'));
      const currentVideo = currentMp4Source.getAttribute('src');

      // Clean up current player resources before loading new content
      currentPlayer.pause();
      currentPlayer.currentTime = 0;
      if (currentPlayer.src) {
        currentPlayer.removeAttribute('src');
        currentPlayer.load(); // Clear existing source
      }
      // Load the current player
      currentPlayer.load();

      // Store the last played video if it's not the transition video
      // currentVideo is encoded from getAttribute('src'), so decode it for storage
      // Also decode transitionVideoPath for comparison since both should be compared unencoded
      if (currentVideo) {
        const parts = currentVideo.split('/');
        const decodedParts = parts.map(part => decodeURIComponent(part));
        const decodedVideo = decodedParts.join('/');
        
        const transitionParts = transitionVideoPath.split('/');
        const decodedTransitionParts = transitionParts.map(part => decodeURIComponent(part));
        const decodedTransitionPath = decodedTransitionParts.join('/');
        
        if (decodedVideo !== decodedTransitionPath) {
          this.playlistManager.setLastPlayed(decodedVideo);
        }
      }

      let videoSrc = this.playlistManager.getLastPlayed() || '';
      // Get the next video if not looping the first video and not in transition mode
      if (!loopFirstVideo && (!transitionVideoPath || !this.isTransition)) {
        videoSrc = this.playlistManager.getNextItem();
        console.debug(`next video: ${videoSrc}`);
      }

      // Set z-index for proper video layering
      // TODO: we can use this opacity to crossfade between mediaFiles
      currentPlayer.style.zIndex = '1';
      currentPlayer.style.opacity = '1';
      nextPlayer.style.zIndex = '0';
      nextPlayer.style.opacity = '0';

      // Handle transition video logic
      if (transitionVideoPath && transitionVideoPath !== '' && this.isTransition) {
        videoSrc = transitionVideoPath; // transitionVideoPath is already encoded
        this.isTransition = false;
      } else {
        this.isTransition = true;
      }

      // Clear the next player's source
      nextMp4Source.src = '';
      nextMp4Source.removeAttribute('src'); // empty source

      // Set the new video source if available
      if (videoSrc) {
        // Clean up next player resources before loading new content
        nextPlayer.pause();
        nextPlayer.currentTime = 0;
        nextMp4Source.setAttribute('src', videoSrc);
        nextPlayer.load();
      }

      // Handle play-once mode
      if (playOnlyOne) {
        // Remove videos after playing once
        currentPlayer.onended = () => {
          currentMp4Source.src = '';
          currentMp4Source.removeAttribute('src');
          nextMp4Source.src = '';
          nextMp4Source.removeAttribute('src');
        };
      }

      // Play the video if it's ready, otherwise wait for it to load
      this.playVideoWhenReady(currentPlayer);

    } catch (error) {
      console.error('Error in playNext function:', error);
    }
  }

  /**
   * Play the video when it's ready
   * @param {HTMLVideoElement} player - The video player element
   */
  playVideoWhenReady(player) {
    const handleLoadedData = () => {
      if (player.readyState >= 3) {
        player.play().catch(err => console.error('Error playing video:', err));
        player.removeEventListener('loadeddata', handleLoadedData, false);
      }
    };

    if (player.readyState >= 3) {
      player.play().catch(err => console.error('Error playing video:', err));
    } else {
      player.addEventListener('loadeddata', handleLoadedData, false);
    }
  }

  /**
   * Initialize the player with the first video
   */
  initialize() {
    const mp4Source = /** @type {HTMLSourceElement} */(this.player1.querySelector('source'));
    let video = this.playlistManager.getNextItem();
    // have to move the state forward after getting the first video
    this.playlistManager.progressPlaylist();

    mp4Source.setAttribute('src', video);
    this.playNext(this.player1, this.player2);
  }
}

/**
 * PlaylistManager handles the playlist state and operations
 */
class PlaylistManager {
  /**
   * @param {string[]} initialPlaylist - The initial list of media files
   * @param {string} hashKey - The hash key for localStorage
   */
  constructor(initialPlaylist, hashKey) {
    /** @type {string[]} */
    this.initialPlaylist = initialPlaylist;
    /** @type {string} */
    this.hashKey = hashKey;
  }

  /**
   * Get the current playlist from localStorage or create a new one
   * @returns {string[]} The current playlist
   */
  getPlaylist() {
    let playlist = [];
    playlist = JSON.parse(localStorage.getItem(`playlist-${this.hashKey}`) || '[]');
    if (!playlist?.length || typeof playlist.pop === 'undefined') {
      playlist = this.createNewPlaylist();
    }
    return playlist;
  }

  /**
   * Create a new shuffled playlist
   * @returns {string[]} A new shuffled playlist
   */
  createNewPlaylist() {
    const playlist = this.shuffleArray(this.initialPlaylist);
    this.storePlaylist(playlist);
    return playlist;
  }

  /**
   * Store the playlist in localStorage
   * @param {string[]} playlist - The playlist to store
   */
  storePlaylist(playlist) {
    localStorage.setItem(`playlist-${this.hashKey}`, JSON.stringify(playlist));
  }

  /**
   * Progress the playlist by removing the last item
   */
  progressPlaylist() {
    const playlist = this.getPlaylist();
    playlist.pop();
    this.storePlaylist(playlist);
  }

  /**
   * Get the next item in the playlist
   * @param {boolean} [returnEncoded=true] - Whether to return the encoded path
   * @returns {string} The next item in the playlist
   */
  getNextItem(returnEncoded = true) {
    let attempts = 0;
    const maxAttempts = this.initialPlaylist.length + 1; // Prevent infinite loops
    
    while (attempts < maxAttempts) {
      const playlist = this.getPlaylist();
      let mediaItem = playlist.pop() || '';
      
      if (!mediaItem) {
        // Playlist is empty, create a new one
        this.createNewPlaylist();
        attempts++;
        continue;
      }

      // Decode mediaItem for comparison since lastPlayed is stored unencoded
      const parts = mediaItem.split('/');
      const decodedParts = parts.map(part => decodeURIComponent(part));
      const decodedMediaItem = decodedParts.join('/');

      console.debug({
        lastPlayed: this.getLastPlayed(),
        mediaItem: decodedMediaItem,
        wasLastPlayed: this.getLastPlayed() === decodedMediaItem
      });

      // check if we played this mediaItem last run (compare unencoded paths)
      if (this.getLastPlayed() === decodedMediaItem) {
        // moves the repeated item to the end so its not skipped entirely
        this.storePlaylist([mediaItem].concat(playlist));
        attempts++;
        continue;
      }

      // Found a valid item that wasn't the last played
      // mediaItem is already pre-encoded in the template, no need for runtime encoding
      if (!returnEncoded) {
        // If we need the unencoded version, decode it
        return decodedMediaItem;
      }
      return mediaItem;
    }

    console.warn('Maximum attempts reached in getNextItem, returning empty string');
    return '';
  }

  /**
   * Set the last played item
   * @param {string} videoSrc - The video source URL
   */
  setLastPlayed(videoSrc) {
    localStorage.setItem(`lastPlayed-${this.hashKey}`, videoSrc);
  }

  /**
   * Get the last played item
   * @returns {string} The last played item
   */
  getLastPlayed() {
    return localStorage.getItem(`lastPlayed-${this.hashKey}`) || '';
  }

  /**
   * Shuffle an array using the Fisher-Yates algorithm
   * @template T
   * @param {T[]} originalArray - The array to shuffle
   * @returns {T[]} The shuffled array
   */
  shuffleArray(originalArray) {
    const array = [...originalArray];
    let currentIndex = array.length;
    let randomIndex, temp;

    // While there remain elements to shuffle
    while (currentIndex > 0) {
      // Pick a remaining element
      randomIndex = Math.floor(Math.random() * currentIndex);
      currentIndex--;

      // And swap it with the current element
      temp = array[currentIndex];
      array[currentIndex] = array[randomIndex];
      array[randomIndex] = temp;
    }

    return array;
  }
}

// Export the classes for use in other files
// @ts-ignore
window.VideoPlayerManager = VideoPlayerManager;
// @ts-ignore
window.PlaylistManager = PlaylistManager;

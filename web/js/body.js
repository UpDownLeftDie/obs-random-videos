// @ts-check

// Initialize the video player manager when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  // Create a new VideoPlayerManager instance
  const playerManager = new VideoPlayerManager('videoPlayer1', 'videoPlayer2');

  // Initialize the player with the first video
  playerManager.initialize();

  // Clean up resources when the page is about to unload
  window.addEventListener('beforeunload', () => {
    if (playerManager && playerManager.cleanup) {
      playerManager.cleanup();
    }
  });

  // Also clean up on page visibility change (when OBS switches scenes)
  document.addEventListener('visibilitychange', () => {
    if (document.hidden && playerManager && playerManager.cleanup) {
      playerManager.cleanup();
    }
  });
});

// @ts-check

// Initialize the video player manager when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  // Create a new VideoPlayerManager instance
  const playerManager = new VideoPlayerManager('videoPlayer1', 'videoPlayer2');

  // Initialize the player with the first video
  playerManager.initialize();
});

const player = document.getElementById('videoPlayer1');
const player2 = document.getElementById('videoPlayer2');
player.addEventListener('ended', () => playNext(player2, player), {
  passive: true,
});
player2.addEventListener('ended', () => playNext(player, player2), {
  passive: true,
});

/* Initial logic */

const mp4Source = player.getElementsByClassName('mp4Source')[0];
let video = `${mediaFiles[0]}`;
// check if we played this video last run
if (mediaFiles.length > 1 && localStorage.getItem('lastPlayed') === video) {
  video = `${mediaFiles[1]}`;
}
mp4Source.setAttribute('src', video);
player.load();

playNext(player, player2);

// @ts-check
const player = /** @type {HTMLMediaElement} */ (
  document.getElementById('videoPlayer1')
);

const player2 = /** @type {HTMLMediaElement} */ (
  document.getElementById('videoPlayer2')
);

if (!isOBS) {
  player.setAttribute('controls', "true");
  player2.setAttribute('controls', "true");
}

player.addEventListener(
  'ended',
  () => {
    if (!loopFirstVideo) {
      progressPlaylistState();
    }
    playNext(player2, player);
  },
  {
    passive: true,
  },
);
player2.addEventListener(
  'ended',
  () => {
    if (!loopFirstVideo) {
      progressPlaylistState();
    }
    playNext(player, player2);
  },
  {
    passive: true,
  },
);

/***** Initial load *****/

const mp4Source = player.getElementsByClassName('mp4Source')[0];
let video = getNextPlaylistItem();
// have to move the state forward after getting the first video
progressPlaylistState();

mp4Source.setAttribute('src', video);

playNext(player, player2);

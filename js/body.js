// @ts-check
const [player1, player2] = document.querySelectorAll('video');
const [mp4Source1] = player1.querySelectorAll('source');
const [mp4Source2] = player2.querySelectorAll('source');

mp4Source1.addEventListener('error', errorFunction);
mp4Source1.addEventListener('stalled', errorFunction);
mp4Source1.addEventListener('suspend', errorFunction);
mp4Source1.addEventListener('waiting', errorFunction);
mp4Source2.addEventListener('error', errorFunction);
mp4Source2.addEventListener('stalled', errorFunction);
mp4Source2.addEventListener('suspend', errorFunction);
mp4Source2.addEventListener('waiting', errorFunction);

/**
 * @param {ErrorEvent | Event} event
 */
function errorFunction(event) {
  const { type, target } = event;
  const videoSrc =
    // @ts-ignore
    /** @type {string} */ (target?.src) || 'UNKNOWN SOURCE FILE ERROR';

  switch (type) {
    case 'error':
      const errorStr = `Error loading: ${videoSrc}`;
      console.error(errorStr);
      if (isOBS) break;

      const errorText = document.createTextNode(errorStr);
      const error = document.createElement('h1');
      error.setAttribute('id', 'error');
      error.appendChild(errorText);
      document.body.innerHTML = '';
      document.body.appendChild(error);
      break;
    case 'stalled':
    case 'suspended':
    case 'waiting':
    default:
      console.error(`Error loading: [TYPE: ${type}] ${videoSrc}`);
  }
}

if (!isOBS) {
  player1.setAttribute('controls', 'true');
  player2.setAttribute('controls', 'true');
}

player1.addEventListener(
  'ended',
  () => {
    if (!loopFirstVideo) {
      progressPlaylistState();
    }
    playNext(player2, player1);
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
    playNext(player1, player2);
  },
  {
    passive: true,
  },
);

/***** Initial load *****/

const mp4Source = player1.getElementsByClassName('mp4Source')[0];
let video = getNextPlaylistItem();
// have to move the state forward after getting the first video
progressPlaylistState();

mp4Source.setAttribute('src', video);

playNext(player1, player2);

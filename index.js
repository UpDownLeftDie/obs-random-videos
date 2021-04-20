const process = require('process');
const fs = require('fs');
const util = require('util');
const inquirer = require('inquirer');
const readdir = util.promisify(fs.readdir);
const template = fs.readFileSync('template.html', 'utf8');

const outputHtmlName = 'obs-random-videos.html';

main();

async function main() {
  const audioFileExts = ['.mp3', '.ogg', '.aac'];
  const videoFileExts = ['.mp4', '.webm', '.mpeg4'];
  const mediaFileExts = videoFileExts.concat(audioFileExts);
  const videoFolder = process.cwd();
  const files = await readdir(videoFolder);

  let mediaFiles = filterFilesExtensions(files, mediaFileExts);
  if (mediaFiles.length < 1) {
    console.error('No media files found!');
    await pressKeyToClose();
  }
  const answers = await askQuestions(mediaFiles);
  mediaFiles = mediaFiles.filter(
    (mediaFile) => mediaFile !== answers.transitionVideo,
  );
  const config = {
    ...answers,
    transitionVideo: `"${answers.transitionVideo}"`,
    videoFolder: `"${videoFolder.replace(/\\/gm, '\\\\')}\\\\"`,
    mediaFiles: JSON.stringify(mediaFiles),
  };

  let finalHtml = template;
  for (const [key, value] of Object.entries(config)) {
    finalHtml = finalHtml.replace(`$${key}$`, value);
  }
  fs.writeFileSync(outputHtmlName, finalHtml);

  console.log(`Output final file too: ${videoFolder}\\${outputHtmlName}`);
  await pressKeyToClose();
}

async function pressKeyToClose() {
  console.log('\nPress any key to close...');
  process.stdin.setRawMode(true);
  return new Promise(() => {
    process.stdin.on('data', () => {
      process.stdin.setRawMode(false);
      process.exit();
    });
  });
}

function filterFilesExtensions(files, extensions) {
  return files.filter((file) => {
    for (let i = 0; i < extensions.length; i++) {
      if (file.endsWith(extensions[i])) return true;
    }
    return false;
  });
}

async function askQuestions(mediaFiles) {
  const questions = [];
  const qPlayOnlyOne = {
    message:
      'Do you only want to play one video? (The first random video will play once and then stop)',
    type: 'confirm',
    name: 'playOnlyOne',
    default: false,
  };
  questions.push(qPlayOnlyOne);

  const qLoopFirstVideo = {
    message: 'Do you want to loop the first video?',
    type: 'confirm',
    name: 'loopFirstVideo',
    default: false,
    when: (a) => !a.playOnlyOne,
  };
  questions.push(qLoopFirstVideo);

  const qTransitionVideoBool = {
    message:
      'Do have a transition video? (This video plays after every other video)',
    type: 'confirm',
    name: 'transitionVideoBool',
    default: false,
    when: (a) => !a.playOnlyOne,
  };
  questions.push(qTransitionVideoBool);

  const qTransitionVideo = {
    message: 'Select your transition video:',
    type: 'list',
    name: 'transitionVideo',
    choices: [...mediaFiles, 'CANCEL'],
    loop: false,
    when: (a) => a.transitionVideoBool,
  };
  questions.push(qTransitionVideo);
  const userAnswers = await inquirer.prompt(questions);
  const answers = {
    playOnlyOne: !!userAnswers.playOnlyOne,
    loopFirstVideo: !!userAnswers.loopFirstVideo,
    transitionVideo:
      !userAnswers.transitionVideo || userAnswers.transitionVideo === 'CANCEL'
        ? ''
        : userAnswers.transitionVideo,
  };
  process.stdin.resume();
  return answers;
}

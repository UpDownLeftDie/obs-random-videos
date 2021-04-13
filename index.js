const fs = require('fs');
const util = require('util');
const inquirer = require('inquirer');
const readdir = util.promisify(fs.readdir);
main();

async function main() {
  const mediaFileExts = ['.mp4', '.mp3'];
  const videoFolder = __dirname;
  const files = await readdir(videoFolder);

  let mediaFiles = filterFilesExtensions(files, mediaFileExts);
  const answers = await askQuestions(mediaFiles);
  mediaFiles = mediaFiles.filter(
    (mediaFile) => mediaFile !== answers.transitionVideo,
  );
  const config = {
    ...answers,
    transitionVideo: `"${answers.transitionVideo}"`,
    videoFolder: `"${videoFolder.replaceAll('\\', '\\\\')}\\\\"`,
    mediaFiles: JSON.stringify(mediaFiles),
  };
  let template = fs.readFileSync('template.html', 'utf8');

  for (const [key, value] of Object.entries(config)) {
    console.log(key, value);
    template = template.replace(`$${key}$`, value);
  }
  fs.writeFileSync('index.html', template);
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
        ? "''"
        : userAnswers.transitionVideo,
  };
  return answers;
}

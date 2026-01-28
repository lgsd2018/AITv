const fs = require('fs');
const https = require('https');
const path = require('path');

const url = "https://image.pollinations.ai/prompt/Goku%20from%20Dragon%20Ball%20and%20Kakashi%20from%20Naruto%20standing%20side%20by%20side%2C%20epic%20composition%2C%20masterpiece%2C%20best%20quality%2C%208K%2C%20ultra%20high%20resolution%2C%20extremely%20detailed%2C%20official%20anime%20artwork%2C%20key%20visual%2C%20cel-shaded%2C%20classic%20cel%20animation%20technique%2C%20flat%20color%20blocks%2C%20crisp%20and%20clean%20line%20art%2C%20bold%20black%20outlines%2C%20hard-edged%20shadows%2C%20no%20gradient%20fills%2C%20vintage%20Japanese%20anime%20aesthetic%2C%20retro%20anime%20style%2C%202D%20hand-drawn%20animation%20style%2C%20soft%20ambient%20lighting%2C%20vibrant%20yet%20cohesive%20colors%2C%20professional%20composition%2C%20clean%20background%2C%20distinct%20color%20layers?width=800&height=450&nologo=true";
const dest = path.join('web', 'src', 'assets', 'images', 'styles', 'cel_shaded_anime.png');

// Ensure directory exists
const dir = path.dirname(dest);
if (!fs.existsSync(dir)){
    fs.mkdirSync(dir, { recursive: true });
}

const file = fs.createWriteStream(dest);
https.get(url, function(response) {
  response.pipe(file);
  file.on('finish', function() {
    file.close(() => console.log('Download completed'));
  });
}).on('error', function(err) {
  fs.unlink(dest, () => {}); // Add empty callback
  console.error('Error downloading image:', err.message);
  process.exit(1);
});

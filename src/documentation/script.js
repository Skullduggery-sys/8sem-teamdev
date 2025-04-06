// Fetch the README.md file
fetch('/README.md')
    .then(response => response.text())
    .then(markdown => {
        // Render the Markdown content into HTML
        const htmlContent = marked.parse(markdown);
        // Insert the rendered HTML into the page
        document.getElementById('readme-content').innerHTML = htmlContent;
    })
    .catch(error => {
        console.error('Error loading README.md:', error);
        document.getElementById('readme-content').innerHTML = '<p>Failed to load README.md.</p>';
    });

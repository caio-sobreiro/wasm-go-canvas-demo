// Minimal JS glue code - just loads and runs the WASM module
(async function() {
    const statusEl = document.getElementById('status');

    try {
        statusEl.textContent = 'Loading WebAssembly module...';

        // Initialize Go WASM runtime
        const go = new Go();

        // Load and instantiate the WASM module
        const result = await WebAssembly.instantiateStreaming(
            fetch('main.wasm'),
            go.importObject
        );

        statusEl.textContent = 'Running! All logic in Go, rendered via Canvas API';

        // Run the Go program (non-blocking)
        go.run(result.instance);

        // Wait for Go to export the animate function
        await new Promise(resolve => setTimeout(resolve, 100));

        // Start animation loop in JS (avoids TinyGo finalizer issues)
        function animate() {
            if (window.goAnimate) {
                window.goAnimate();
            }
            requestAnimationFrame(animate);
        }
        animate();

    } catch (err) {
        statusEl.textContent = 'Error: ' + err.message;
        statusEl.style.color = '#f44336';
        console.error('Failed to load WASM:', err);
    }
})();

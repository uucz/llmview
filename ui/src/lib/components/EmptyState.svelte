<div class="empty">
  <div class="empty-icon">
    <div class="icon-box"></div>
  </div>
  <h2>Waiting for LLM traffic</h2>
  <p>Point your LLM client through llmview to get started.</p>
  <div class="commands">
    <div class="cmd">
      <span class="cmd-label">OpenAI</span>
      <code>export OPENAI_BASE_URL=http://localhost:4700/proxy/openai/v1</code>
    </div>
    <div class="cmd">
      <span class="cmd-label">Anthropic</span>
      <code>export ANTHROPIC_BASE_URL=http://localhost:4700/proxy/anthropic</code>
    </div>
    <div class="cmd">
      <span class="cmd-label">Ollama</span>
      <code>export OLLAMA_HOST=http://localhost:4700/proxy/ollama</code>
    </div>
  </div>
</div>

<style>
  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 24px;
    text-align: center;
    animation: fadeUp 0.5s ease both;
  }

  .empty-icon {
    margin-bottom: 24px;
    animation: float 4s ease-in-out infinite;
  }

  .icon-box {
    width: 48px;
    height: 48px;
    border-radius: var(--radius);
    background: var(--accent-dim);
    border: 1px solid color-mix(in srgb, var(--accent) 15%, transparent);
  }

  h2 {
    font-family: var(--font-sans);
    font-size: 18px;
    font-weight: 600;
    color: var(--text-1);
    margin-bottom: 6px;
    letter-spacing: -0.01em;
  }

  p {
    color: var(--text-2);
    font-size: 13px;
    margin-bottom: 32px;
  }

  .commands {
    display: flex;
    flex-direction: column;
    gap: 6px;
    width: 100%;
    max-width: 560px;
  }

  .cmd {
    display: flex;
    align-items: center;
    gap: 12px;
    text-align: left;
  }

  .cmd-label {
    font-family: var(--font-sans);
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--text-2);
    min-width: 72px;
    text-align: right;
    font-weight: 600;
  }

  code {
    flex: 1;
    background: var(--surface-2);
    padding: 9px 14px;
    border-radius: var(--radius-sm);
    color: var(--accent);
    font-size: 12px;
    border: 1px solid var(--border);
    user-select: all;
    cursor: text;
    transition: border-color 0.2s;
  }

  code:hover {
    border-color: var(--border-hover);
  }
</style>

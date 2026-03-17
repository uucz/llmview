<div class="empty">
  <div class="empty-icon">
    <!-- Terminal icon — line-art -->
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="48" height="48">
      <polyline points="4 17 10 11 4 5" />
      <line x1="12" y1="19" x2="20" y2="19" />
    </svg>
  </div>
  <h2>Waiting for LLM traffic</h2>
  <p>Point your LLM client through llmview to start monitoring.</p>
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
    margin-bottom: 20px;
    color: var(--text-tertiary);
    opacity: 0.25;
    animation: float 4s ease-in-out infinite;
  }

  h2 {
    font-family: var(--font-heading);
    font-size: 18px;
    font-weight: 600;
    color: var(--text-secondary);
    margin-bottom: 6px;
    letter-spacing: -0.01em;
  }

  p {
    font-family: var(--font-body);
    color: var(--text-tertiary);
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
    font-family: var(--font-heading);
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--text-tertiary);
    min-width: 72px;
    text-align: right;
    font-weight: 600;
  }

  code {
    flex: 1;
    font-family: var(--font-mono);
    background: var(--surface-2);
    padding: 9px 14px;
    border-radius: var(--radius-sm);
    color: var(--brand-orange);
    font-size: 12px;
    border: 1px solid var(--border-color);
    user-select: all;
    cursor: text;
    transition: border-color 0.2s;
  }

  code:hover {
    border-color: rgba(217, 119, 87, 0.2);
  }
</style>

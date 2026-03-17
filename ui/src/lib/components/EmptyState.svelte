<div class="empty">
  <div class="empty-icon">◉</div>
  <h2>Waiting for LLM traffic...</h2>
  <p>Point your LLM client through llmview:</p>
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
  }

  .empty-icon {
    font-size: 48px;
    color: var(--accent);
    margin-bottom: 16px;
    opacity: 0.3;
    animation: float 3s ease-in-out infinite;
  }

  @keyframes float {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-8px); }
  }

  h2 {
    font-size: 18px;
    font-weight: 600;
    color: var(--text-2);
    margin-bottom: 8px;
    font-family: var(--font-sans);
  }

  p {
    color: var(--text-2);
    font-size: 13px;
    margin-bottom: 24px;
  }

  .commands {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
    max-width: 540px;
  }

  .cmd {
    display: flex;
    align-items: center;
    gap: 12px;
    text-align: left;
  }

  .cmd-label {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-2);
    min-width: 70px;
    text-align: right;
    font-weight: 600;
  }

  code {
    flex: 1;
    background: var(--bg-2);
    padding: 8px 14px;
    border-radius: var(--radius-sm);
    color: var(--accent);
    font-size: 12px;
    border: 1px solid var(--border);
    user-select: all;
    cursor: text;
  }
</style>

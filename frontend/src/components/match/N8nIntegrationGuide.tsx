import { useState } from "react";
import { Button } from "../ui/Button";

export function N8nIntegrationGuide() {
    const [expanded, setExpanded] = useState(false);

    return (
        <div className="border border-purple-200 bg-purple-50 rounded p-4 mb-6">
            <button
                onClick={() => setExpanded(!expanded)}
                className="flex items-center justify-between w-full"
            >
                <div className="flex items-center gap-2">
                    <span className="text-xl">⚙️</span>
                    <h3 className="font-semibold text-purple-900">N8N Integration</h3>
                </div>
                <span className="text-purple-600">{expanded ? "−" : "+"}</span>
            </button>

            {expanded && (
                <div className="mt-4 space-y-4 text-sm text-purple-800">
                    <p>
                        <strong>Enable automated screenshot processing:</strong>
                    </p>

                    <div className="bg-white rounded p-3">
                        <p className="text-xs font-mono bg-gray-100 p-2 rounded mb-2">
                            Webhook endpoint: https://yourn8ninstance.com/webhook/tourney-screenshot
                        </p>
                        <ol className="list-decimal list-inside space-y-1 text-xs">
                            <li>Set up n8n workflow with Vision API</li>
                            <li>Take a screenshot (Ctrl+Shift+S)</li>
                            <li>Send to n8n webhook</li>
                            <li>Receive processed screenshot URL + stats</li>
                            <li>Paste URL below and stats auto-fill</li>
                        </ol>
                    </div>

                    <div className="border-t border-purple-200 pt-3">
                        <p className="text-xs font-semibold mb-2">Testing without N8N:</p>
                        <p className="text-xs">
                            Use placeholder: <code className="bg-gray-100 px-1">
                                https://via.placeholder.com/800x600
                            </code>
                        </p>
                    </div>

                    <Button
                        size="sm"
                        onClick={() =>
                            window.open("/docs/N8N_INTEGRATION.md", "_blank")
                        }
                        className="text-xs"
                    >
                        View Full Integration Guide →
                    </Button>
                </div>
            )}
        </div>
    );
}

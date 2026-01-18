/**
 * Voice Commands - STUB for minimal voice system
 *
 * This is a minimal stub to prevent compilation errors.
 * The minimal voice system doesn't parse commands - it's just simple chat.
 * This stub ensures existing code compiles but command parsing is effectively disabled.
 */

export interface VoiceCommand {
	type: 'unknown';
	confidence: number;
	raw: string;
	module?: string;
	view?: string;
	app?: string;
	nodeId?: string;
}

export const voiceCommandParser = {
	parse: (text: string): VoiceCommand => {
		// Always return 'unknown' - no command parsing in minimal system
		return {
			type: 'unknown',
			confidence: 0,
			raw: text
		};
	}
};

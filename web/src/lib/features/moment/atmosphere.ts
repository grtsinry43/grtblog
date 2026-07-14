import type { MomentAtmosphere, MomentMood, MomentWeather } from './types';
import type { LucideIconKey } from '$lib/ui/icons/lucide-loaders';

export type MomentAtmosphereDisplayItem = {
	kind: 'weather' | 'mood';
	label: string;
	icon: LucideIconKey;
};

const weatherConfig: Record<MomentWeather, Omit<MomentAtmosphereDisplayItem, 'kind'>> = {
	sunny: { label: '晴朗', icon: 'sun' },
	cloudy: { label: '多云', icon: 'cloud-sun' },
	overcast: { label: '阴天', icon: 'cloud' },
	rainy: { label: '下雨', icon: 'cloud-rain' },
	snowy: { label: '下雪', icon: 'cloud-snow' },
	windy: { label: '有风', icon: 'wind' },
	foggy: { label: '有雾', icon: 'cloud-fog' }
};

const moodConfig: Record<MomentMood, Omit<MomentAtmosphereDisplayItem, 'kind'>> = {
	joyful: { label: '开心', icon: 'smile' },
	calm: { label: '平静', icon: 'heart' },
	excited: { label: '兴奋', icon: 'sparkles' },
	tired: { label: '疲惫', icon: 'moon-star' },
	sad: { label: '低落', icon: 'frown' }
};

const isMomentWeather = (value: unknown): value is MomentWeather =>
	typeof value === 'string' && Object.hasOwn(weatherConfig, value);

const isMomentMood = (value: unknown): value is MomentMood =>
	typeof value === 'string' && Object.hasOwn(moodConfig, value);

export const getMomentAtmosphereDisplayItems = (
	atmosphere?: MomentAtmosphere | null
): MomentAtmosphereDisplayItem[] => {
	const items: MomentAtmosphereDisplayItem[] = [];
	if (isMomentWeather(atmosphere?.weather)) {
		items.push({ kind: 'weather', ...weatherConfig[atmosphere.weather] });
	}
	if (isMomentMood(atmosphere?.mood)) {
		items.push({ kind: 'mood', ...moodConfig[atmosphere.mood] });
	}
	return items;
};

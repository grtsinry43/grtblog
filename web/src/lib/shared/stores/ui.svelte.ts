export class GlobalUIState {
	isSearchOpen = $state(false);

	toggleSearch() {
		this.isSearchOpen = !this.isSearchOpen;
	}

	openSearch() {
		this.isSearchOpen = true;
	}

	closeSearch() {
		this.isSearchOpen = false;
	}
}

export const uiState = new GlobalUIState();

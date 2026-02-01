import { createModelDataContext } from "svatoms";
import type { AuthApproachState } from "./types";

export const AuthCtx = createModelDataContext<AuthApproachState>({
    name: "auth",
    initial: {
        turnstile: {
            enabled: false,
            siteKey: "",
            error: ""
        },
        oauth: {
            providers: [],
            error: "",
            loadingKey: null
        },
        login: {
            loading: false,
            error: ""
        },
        showPasswordLogin: false
    }
});

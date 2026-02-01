export namespace main {
	
	export class Config {
	    frontendHost: string;
	    frontendPort: number;
	    backendHost: string;
	    backendPort: number;
	    username: string;
	    password: string;
	    dataPath: string;
	    backupPath: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.frontendHost = source["frontendHost"];
	        this.frontendPort = source["frontendPort"];
	        this.backendHost = source["backendHost"];
	        this.backendPort = source["backendPort"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.dataPath = source["dataPath"];
	        this.backupPath = source["backupPath"];
	    }
	}
	export class StartBackendServerResult {
	    generated: boolean;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new StartBackendServerResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generated = source["generated"];
	        this.password = source["password"];
	    }
	}

}


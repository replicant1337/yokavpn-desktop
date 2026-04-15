export namespace domain {
	
	export class PingResult {
	    index: number;
	    name: string;
	    latency_ms: number;
	    success: boolean;
	    error: string;
	    target: string;
	
	    static createFrom(source: any = {}) {
	        return new PingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.name = source["name"];
	        this.latency_ms = source["latency_ms"];
	        this.success = source["success"];
	        this.error = source["error"];
	        this.target = source["target"];
	    }
	}
	export class Stats {
	    upload_bytes: number;
	    download_bytes: number;
	    upload_rate: number;
	    download_rate: number;
	    connections: number;
	
	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.upload_bytes = source["upload_bytes"];
	        this.download_bytes = source["download_bytes"];
	        this.upload_rate = source["upload_rate"];
	        this.download_rate = source["download_rate"];
	        this.connections = source["connections"];
	    }
	}

}


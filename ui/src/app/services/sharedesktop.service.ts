import { EventEmitter, Injectable, Output } from '@angular/core';
import { Config } from 'src/app/config/config';



@Injectable({
    providedIn: 'root',
})
export class ShareDesktopService {

    isactive: boolean = false;
    sharingLink: string = "";

    async enableSharing(viewOnly: boolean): Promise<number> {
        var data = {
            "secret": localStorage.getItem('secret'),
            "action": "activate",
            "view_only": viewOnly
        }

        let response = fetch(Config.ApiEndpoint + "/desktop/share", {
            "method": "POST",
            "body": JSON.stringify(data)
        })

        let rdata = await response

        switch (rdata.status) {
            case 200:
                let respBody = await rdata.json()
                this.isactive = true
                this.sharingLink = document.location.origin + "/#/shared-desktop/" + respBody.share_secret + ":" + respBody.client_id
                return 0
            case 410:   // desktop not active
                this.isactive = false
                return 1
        }

        this.isactive = false
        return -1
    }

    async disableSharing(): Promise<number> {
        var data = {
            "secret": localStorage.getItem('secret'),
            "action": "deactivate",
        }

        let response = fetch(Config.ApiEndpoint + "/desktop/share", {
            "method": "POST",
            "body": JSON.stringify(data)
        })

        let rdata = await response

        if (!rdata.ok) {
            return -1
        }

        if (rdata.status == 200) {
            this.isactive = false
            return 0
        }
        return 1
    }

}
#include <radosstriper/libradosstriper.h>
#include <errno.h>
#include <stdlib.h>

int try_break_lock(rados_ioctx_t io_ctx, rados_striper_t striper, char *oid){
        int exclusive;
        char tag[1024];
        char clients[1024];
        char cookies[1024];
        char addresses[1024];
        int ret = 0;
        size_t tag_len = 1024;
        size_t clients_len = 1024;
        size_t cookies_len = 1024;
        size_t addresses_len = 1024;
        char *firstObjOid = NULL;
        char tail[] = ".0000000000000000";
        firstObjOid = (char*)malloc(strlen(oid)+strlen(tail)+1);
        strcpy(firstObjOid, oid);
        strcat(firstObjOid, tail);
        ret = rados_list_lockers(io_ctx, firstObjOid, "striper.lock", &exclusive, tag, &tag_len, clients, &clients_len, cookies, &cookies_len, addresses, &addresses_len);
        free(firstObjOid);
        return ret;
}


/* force delete the file */
int striprados_remove(rados_ioctx_t io_ctx, rados_striper_t striper, char *oid){
        int ret;
        int retry = 0;
retry:
        ret = rados_striper_remove(striper, oid);
        if (ret == -EBUSY && retry == 0){
                ret = try_break_lock(io_ctx,striper,oid);
                retry++;
                if (ret == 0)
                        goto retry;
        }
}

/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

#ifndef _GSE_DATA_PROXY_EXPORTER_GSEDATA_PACKAGE_H_
#define _GSE_DATA_PROXY_EXPORTER_GSEDATA_PACKAGE_H_

#include <string.h>
#include <stdint.h>

#include <string>
#include <vector>

#include "open/protocol_head.h"

namespace gse { 
namespace dataserver {

#ifndef DS_MAGIC_NUM
#define DS_MAGIC_NUM   0xCAFECAFE
#endif


class GSEDataPackage
{
public:
    enum
    {
       GSE_DATA_PUB_HEAD_SIZE = sizeof(DataMsgHead),
       GSE_DATA_MESSAGE_BLOCK_SIZE = (1024 + 512) // default: each message block size is 1526B
    };
public:
    GSEDataPackage();
    ~GSEDataPackage();

public:
    void ResetBufferPosition();
    char* GetDataPointer();
    uint32_t GetDataLength();

    void SetId(std::string &id);
    void SetChannelId(uint32_t channelid);
    void SetExterntions(std::vector<std::string> &externtions);
    void Pack(const char* ptr_data, uint32_t data_len);
    void SetCreateTimestamp(uint32_t timestamp);
    void SetArrivedTimestamp(uint32_t timestamp);
private:
    void tryReallocBuffer(uint32_t targetSize);

    int CalcMsgLen(uint32_t data_len);
private:
    //DISALLOW_COPY_AND_ASSIGN(GSEDataPackage);


private:
    char*                     m_ptrValue;
    uint32_t                  m_valueLength;
    uint32_t                  m_valuePosition;
    uint32_t                  m_createTimestamp;
    uint32_t                  m_arrivedTimeStamp;

    uint32_t                m_msgLen;
    std::string m_id;
    uint32_t m_channelId;
    std::vector<std::string> m_externtions;
};

}
}
#endif

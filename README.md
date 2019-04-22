# Gecko 概念

    :input state
    | ByteEncoder         | 
    | InputDevice (#uuid) |
    | MesgDecoder         |
    
            /|\
             |
            \|/
            
    :intercept state
    | Interceptors (#topic) |
    
            /|\
             |
            \|/
            
    :drive state
    | Drivers (#topic) | Triggers (#topic) |
    
            /|\                 |
             |                  |          
            \|/                \|/
    | MesgDecoder      |
    | Outputs (#uuid)  | Outputs (#uuid) |
    | ByteEncoder      |
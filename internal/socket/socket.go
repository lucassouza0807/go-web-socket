package socket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type        string    `json:"type"`
	To          string    `json:"to"`
	Message     string    `json:"message"`
	From        string    `json:"from"`
	Status      string    `json:"data"`
	Timestamp   time.Time `json:"timestamp"`
	FileId      string    `json:"fileId,omitempty"`
	ChunkIndex  int       `json:"chunkIndex,omitempty"`
	TotalChunks int       `json:"totalChunks,omitempty"`
	ChunkData   string    `json:"chunkData,omitempty"`
	MediaType   string    `json:"media_type"`
	MimeType    string    `json:"mime_type"`
	Filename    string    `json:"filename"`
	FileUrl     string    `json:"fileurl"`
}

var (
	clients    sync.Map
	fileChunks = make(map[string]map[int][]byte)
	chunkMutex sync.Mutex
	upgrader   = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

// 📌 Envia mensagem via HTTP (REST API)
func SendMessage(ctx *gin.Context) {
	var msg Message
	msg.Timestamp = time.Now()

	// Decodifica JSON recebido
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "JSON inválido"})
		return
	}

	// Converte a mensagem para JSON
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Erro ao processar mensagem"})
		return
	}

	// Se for mensagem privada, envia direto ao destinatário
	if msg.Type == "private" {
		if err := sendPrivateMessage(msg.To, messageBytes); err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "Usuário não encontrado"})
			return
		}
	} else {
		// Broadcast para todos os clientes
		broadcastMessage(messageBytes)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Mensagem enviada com sucesso",
		"data":    msg,
	})
}

// 📌 Retorna os usuários online
func GetOnlineUsers(ctx *gin.Context) {
	var users []string
	clients.Range(func(key, _ interface{}) bool {
		users = append(users, key.(string))
		return true
	})

	ctx.JSON(http.StatusOK, gin.H{"online_users": users})
}

// 📌 Manipula conexões WebSocket
func HandleSocket(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	if userID == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "user_id é obrigatório"})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Erro ao estabelecer WebSocket", "details": err.Error()})
		return
	}

	defer conn.Close()
	clients.Store(userID, conn)
	broadcastUserStatus(userID, true)

	fmt.Println("Novo usuário conectado:", userID)

	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("Usuário desconectado:", userID)
		clients.Delete(userID)
		broadcastUserStatus(userID, false)
		return nil
	})

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if messageType == websocket.TextMessage {
			var msgData Message
			if err := json.Unmarshal(message, &msgData); err != nil {
				continue
			}

			msgData.From = userID

			switch msgData.Type {
			case "private":
				if err := sendPrivateMessage(msgData.To, message); err != nil {
					conn.WriteMessage(websocket.TextMessage, []byte("Erro: "+err.Error()))
				}
			default:
				broadcastMessage(message)
			}
		} else if messageType == websocket.BinaryMessage {
			err := handleFileChunk(userID, message)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte("Erro ao processar o arquivo: "+err.Error()))
			} else {
				conn.WriteMessage(websocket.TextMessage, []byte("Chunk de arquivo recebido"))
			}
		}
	}

	clients.Delete(userID)
	broadcastUserStatus(userID, false)
	fmt.Println("Usuário desconectado:", userID)
}

// 📌 Manipula os chunks de arquivo recebidos
func handleFileChunk(userID string, message []byte) error {
	fmt.Println(string(message))

	var msg Message

	if err := json.Unmarshal(message, &msg); err != nil {
		return err
	}

	chunkData, err := decodeBase64(msg.ChunkData)
	if err != nil {
		return err
	}

	chunkMutex.Lock()
	defer chunkMutex.Unlock()

	if _, exists := fileChunks[msg.FileId]; !exists {
		fileChunks[msg.FileId] = make(map[int][]byte)
	}

	fileChunks[msg.FileId][msg.ChunkIndex] = chunkData
	fmt.Printf("Recebido chunk %d de %d para arquivo %s\n", msg.ChunkIndex+1, msg.TotalChunks, msg.FileId)

	if len(fileChunks[msg.FileId]) == msg.TotalChunks {
		return finalizeFileUpload(msg.FileId)
	}

	return nil
}

// 📌 Decodifica dados base64 recebidos do WebSocket
func decodeBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// 📌 Finaliza a reconstrução do arquivo
func finalizeFileUpload(fileId string) error {
	filePath := fmt.Sprintf("uploads/%s_reconstructed", fileId)

	if _, err := os.Stat("uploads/"); os.IsNotExist(err) {
		os.Mkdir("uploads/", 0755)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := 0; i < len(fileChunks[fileId]); i++ {
		if _, err := file.Write(fileChunks[fileId][i]); err != nil {
			return err
		}
	}

	fmt.Println("Arquivo reconstruído com sucesso:", filePath)
	delete(fileChunks, fileId)
	return nil
}

// 📌 Enviar mensagem privada
func sendPrivateMessage(toUser string, message []byte) error {
	if conn, ok := clients.Load(toUser); ok {
		return conn.(*websocket.Conn).WriteMessage(websocket.TextMessage, message)
	}
	return fmt.Errorf("usuário não encontrado")
}

// 📌 Broadcast para todos os clientes conectados
func broadcastMessage(message []byte) {
	clients.Range(func(_, clientConn interface{}) bool {
		clientConn.(*websocket.Conn).WriteMessage(websocket.TextMessage, message)
		return true
	})
}

// 📌 Notifica usuários sobre conexão/desconexão
func broadcastUserStatus(userID string, connected bool) {
	status := "user-disconnected"
	if connected {
		status = "user-connected"
	}

	msg := Message{
		Type:   "status",
		From:   userID,
		Status: status,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Erro ao serializar mensagem de status:", err)
		return
	}

	broadcastMessage(msgBytes)
}

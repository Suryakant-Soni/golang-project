// this package will be used for creating protobuff server and apis
package grpcapi

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mailinglist/mdb"
	pb "mailinglist/proto"
	"net"
	"time"

	"google.golang.org/grpc"
)

type MailServer struct {
	pb.UnimplementedMailingListServiceServer
	db *sql.DB
}

// to convert protobuff based struc of mail.proto to simple go based struct in mdb go file
func pbEntryToMdbEntry(pbEntry *pb.EmailEntry) mdb.EmailEntry {
	t := time.Unix(pbEntry.ConfirmedAt, 0)
	return mdb.EmailEntry{
		Id:          pbEntry.Id,
		Email:       pbEntry.Email,
		ConfirmedAt: &t,
		OptOut:      pbEntry.OptOut,
	}
}

func mdbEntryToPbEntry(mdbEntry *mdb.EmailEntry) pb.EmailEntry {
	return pb.EmailEntry{
		Id:          mdbEntry.Id,
		Email:       mdbEntry.Email,
		ConfirmedAt: mdbEntry.ConfirmedAt.Unix(),
		OptOut:      mdbEntry.OptOut,
	}
}

// a utility function which will be used to get email in grpc pb format from the database mdb
func EmailResponse(db *sql.DB, email string) (*pb.EmailResponse, error) {
	entry, err := mdb.GetEmail(db, email)
	log.Println("entry from mdb", entry)
	if err != nil {
		return &pb.EmailResponse{}, err
	}
	if entry == nil {
		return &pb.EmailResponse{}, nil
	}
	log.Println("recahed line 52")
	res := mdbEntryToPbEntry(entry)
	log.Println("entry in pb format", res)
	return &pb.EmailResponse{EmailEntry: &res}, nil
}

func (s *MailServer) GetEmail(ctx context.Context, req *pb.GetEmailRequest) (*pb.EmailResponse, error) {
	log.Printf("GetEmail req param: %v \n", req)
	return EmailResponse(s.db, req.EmailAddr)
}

func (s *MailServer) GetEmailBatch(ctx context.Context, req *pb.GetEmailBatchRequest) (*pb.GetEmailBatchResponse, error) {
	log.Printf("grpc GetEmailBatch: %v \n", req)
	// created a parameter for batch info to be sent to mdb method for getting data in form of batch
	params := mdb.GetEmailBatchQueryParams{
		Page:  int(req.Page),
		Count: int(req.Count),
	}
	// get emails from mdb in slice
	mdbEntries, err := mdb.GetEmailBatch(s.db, params)
	if err != nil {
		return &pb.GetEmailBatchResponse{}, err
	}
	pbEntries := make([]*pb.EmailEntry, 0, len(mdbEntries))
	// converting slice of mdb into mdb format
	for i := 0; i < len(mdbEntries); i++ {
		entry := mdbEntryToPbEntry(&mdbEntries[i])
		pbEntries = append(pbEntries, &entry)
	}
	return &pb.GetEmailBatchResponse{EmailEntries: pbEntries}, nil
}

func (s *MailServer) CreateEmail(ctx context.Context, req *pb.CreateEmailRequest) (*pb.EmailResponse, error) {
	log.Printf("grpc CreateEmail: %v \n", req)
	err := mdb.CreateEmail(s.db, req.EmailAddr)
	if err != nil {
		return &pb.EmailResponse{}, err
	}
	return EmailResponse(s.db, req.EmailAddr)
}

func (s *MailServer) UpdateEmail(ctx context.Context, req *pb.UpdateEmailRequest) (*pb.EmailResponse, error) {
	log.Printf("grpc UpdateEmail: %v \n", req)
	entry := pbEntryToMdbEntry(req.EmailEntry)
	err := mdb.UpdateEmail(s.db, entry)
	if err != nil {
		return &pb.EmailResponse{}, err
	}
	return EmailResponse(s.db, entry.Email)
}

func (s *MailServer) DeleteEmail(ctx context.Context, req *pb.DeleteEmailRequest) (*pb.EmailResponse, error) {
	log.Printf("grpc DeleteEmail: %v \n", req)
	err := mdb.DeleteEmail(s.db, req.EmailAddr)
	if err != nil {
		return &pb.EmailResponse{}, err
	}
	return EmailResponse(s.db, req.EmailAddr)
}

func Serve(db *sql.DB, bind string) {
	// it will announce or bind the port and the protocol to be used
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatalf("gRPC server error : failure to bind")
	}
	// create a new grpc server instance
	grpcServer := grpc.NewServer()
	// create an instance of mailServer type
	mailServer := MailServer{db: db}

	// register handlers of mailinglist service awhich are attached to mailServer type with the grpc server
	// using a generated method
	pb.RegisterMailingListServiceServer(grpcServer, &mailServer)
	fmt.Printf("grpc api server listening on %v \n", bind)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("grpc server error %v \n", err)
	}
}
